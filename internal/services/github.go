package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/datmedevil17/go-vuln/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GitHubService struct {
	db *gorm.DB
}

type GitHubRepo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Language    string `json:"language"`
	Private     bool   `json:"private"`
}

type GitHubFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type GitHubIssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type GitHubIssue struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{db: db}
}

// ListRepositories fetches repositories from GitHub API
func (s *GitHubService) ListRepositories(ctx context.Context, accessToken string, userID uuid.UUID) ([]models.Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?per_page=100", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	var githubRepos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&githubRepos); err != nil {
		return nil, err
	}

	// Convert and store in database
	var repositories []models.Repository
	for _, gr := range githubRepos {
		repo := models.Repository{
			UserID:      userID,
			GitHubID:    gr.ID,
			FullName:    gr.FullName,
			Name:        gr.Name,
			Description: gr.Description,
			HTMLURL:     gr.HTMLURL,
			Language:    gr.Language,
			IsPrivate:   gr.Private,
		}

		// Upsert repository
		var existingRepo models.Repository
		result := s.db.Where("git_hub_id = ?", gr.ID).First(&existingRepo)
		if result.Error == gorm.ErrRecordNotFound {
			s.db.Create(&repo)
		} else {
			s.db.Model(&existingRepo).Updates(repo)
			repo = existingRepo
		}

		repositories = append(repositories, repo)
	}

	return repositories, nil
}

// GetRepositoryFiles fetches file tree from GitHub
func (s *GitHubService) GetRepositoryFiles(ctx context.Context, accessToken, owner, repo, path string) ([]GitHubFile, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch files: %s", resp.Status)
	}

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

// GetFileContent fetches content of a specific file
func (s *GitHubService) GetFileContent(ctx context.Context, accessToken, owner, repo, path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// CreateIssue creates a new issue in a GitHub repository
func (s *GitHubService) CreateIssue(ctx context.Context, accessToken, owner, repo, title, body string) (*GitHubIssue, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)

	issueReq := GitHubIssueRequest{
		Title: title,
		Body:  body,
	}

	jsonData, err := json.Marshal(issueReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create issue: %s - %s", resp.Status, string(body))
	}

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// Auto-Fix Structs

type CreateBranchRequest struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

type UpdateFileRequest struct {
	Message string `json:"message"`
	Content string `json:"content"`
	Sha     string `json:"sha"`
	Branch  string `json:"branch"`
}

type CreatePullRequestRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type GitHubRef struct {
	Ref    string `json:"ref"`
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
}

type GitHubPR struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
}

// Methods

// GetReference fetches a git reference (e.g. heads/main)
func (s *GitHubService) GetReference(ctx context.Context, accessToken, owner, repo, ref string) (*GitHubRef, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/%s", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get ref: %s", resp.Status)
	}

	var gitRef GitHubRef
	if err := json.NewDecoder(resp.Body).Decode(&gitRef); err != nil {
		return nil, err
	}
	return &gitRef, nil
}

// CreateBranch creates a new branch
func (s *GitHubService) CreateBranch(ctx context.Context, accessToken, owner, repo, newBranch, baseSha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	bodyReq := CreateBranchRequest{
		Ref: "refs/heads/" + newBranch,
		Sha: baseSha,
	}
	jsonData, _ := json.Marshal(bodyReq)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create branch: %s - %s", resp.Status, string(body))
	}
	return nil
}

// GetFileSHA fetches the SHA of a file
func (s *GitHubService) GetFileSHA(ctx context.Context, accessToken, owner, repo, path, branch string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get file sha: %s", resp.Status)
	}

	// Wait, GitHubFile doesn't have SHA field. Need to check if I can add it or use map.
	// Let's use a temporary struct or map.
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if sha, ok := result["sha"].(string); ok {
		return sha, nil
	}
	return "", fmt.Errorf("sha not found in response")
}

// UpdateFile updates (commits) a file
func (s *GitHubService) UpdateFile(ctx context.Context, accessToken, owner, repo, path, content, sha, message, branch string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	bodyReq := UpdateFileRequest{
		Message: message,
		Content: content, // Must be base64 encoded? GitHub API expects base64 unless using raw accept header for reading. For writing, struct `content` usually needs base64.
		Sha:     sha,
		Branch:  branch,
	}
	// Note: content must be base64 encoded.

	jsonData, _ := json.Marshal(bodyReq)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update file: %s - %s", resp.Status, string(body))
	}
	return nil
}

// CreatePullRequest creates a PR
func (s *GitHubService) CreatePullRequest(ctx context.Context, accessToken, owner, repo, title, body, head, base string) (*GitHubPR, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	bodyReq := CreatePullRequestRequest{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	}

	jsonData, _ := json.Marshal(bodyReq)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create PR: %s - %s", resp.Status, string(body))
	}

	var pr GitHubPR
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
