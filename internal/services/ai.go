package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/datmedevil17/go-vuln/internal/config"
)

type AIService struct {
	config *config.Config
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{config: cfg}
}

// AnalyzeCode uses AI to analyze code for vulnerabilities
func (s *AIService) AnalyzeCode(ctx context.Context, code string, language string) (string, error) {
	prompt := fmt.Sprintf(`Analyze the following %s code for security vulnerabilities. 
Provide a detailed security analysis including:
1. Identified vulnerabilities
2. Severity level (Critical, High, Medium, Low)
3. Detailed explanation
4. Recommended fixes

Code:
%s`, language, code)

	// Try Gemini first, fallback to Groq
	if s.config.AI.GeminiAPIKey != "" {
		result, err := s.callGemini(ctx, prompt)
		if err == nil {
			return result, nil
		}
	}

	if s.config.AI.GroqAPIKey != "" {
		return s.callGroq(ctx, prompt)
	}

	return "", fmt.Errorf("no AI API keys configured")
}

// GenerateSecurityRecommendations generates security recommendations
func (s *AIService) GenerateSecurityRecommendations(ctx context.Context, scanResults string) (string, error) {
	prompt := fmt.Sprintf(`Based on the following security scan results, provide detailed security recommendations:

Scan Results:
%s

Please provide:
1. Priority recommendations
2. Quick wins (easy to implement)
3. Long-term security improvements
4. Best practices to follow`, scanResults)

	if s.config.AI.GeminiAPIKey != "" {
		result, err := s.callGemini(ctx, prompt)
		if err == nil {
			return result, nil
		}
	}

	if s.config.AI.GroqAPIKey != "" {
		return s.callGroq(ctx, prompt)
	}

	return "", fmt.Errorf("no AI API keys configured")
}

// GenerateFix generates a fix for vulnerable code
func (s *AIService) GenerateFix(ctx context.Context, code string, vulnerability string) (string, error) {
	prompt := fmt.Sprintf(`You are a security expert. Fix the following code to resolve the specified vulnerability.
Return ONLY the fixed code without any markdown formatting or explanation.

Vulnerability: %s

Code:
%s`, vulnerability, code)

	if s.config.AI.GeminiAPIKey != "" {
		result, err := s.callGemini(ctx, prompt)
		if err == nil {
			return result, nil
		}
	}

	if s.config.AI.GroqAPIKey != "" {
		return s.callGroq(ctx, prompt)
	}

	return "", fmt.Errorf("no AI API keys configured")
}

// ChatResponse generates a chatbot response
func (s *AIService) ChatResponse(ctx context.Context, userMessage string, conversationHistory []map[string]string) (string, error) {
	prompt := "You are a cybersecurity expert assistant. Help users understand security vulnerabilities and provide guidance.\n\n"

	// Add conversation history
	for _, msg := range conversationHistory {
		prompt += fmt.Sprintf("%s: %s\n", msg["role"], msg["content"])
	}
	prompt += fmt.Sprintf("User: %s\nAssistant:", userMessage)

	if s.config.AI.GroqAPIKey != "" {
		return s.callGroq(ctx, prompt)
	}

	if s.config.AI.GeminiAPIKey != "" {
		return s.callGemini(ctx, prompt)
	}

	return "", fmt.Errorf("no AI API keys configured")
}

// callGemini makes a request to Google Gemini API
func (s *AIService) callGemini(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", s.config.AI.GeminiAPIKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %s - %s", resp.Status, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from Gemini")
}

// callGroq makes a request to Groq API
func (s *AIService) callGroq(ctx context.Context, prompt string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	reqBody := GroqRequest{
		Model: "llama-3.3-70b-versatile",
		Messages: []GroqMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.GroqAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Groq API error: %s - %s", resp.Status, string(body))
	}

	var groqResp GroqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", err
	}

	if len(groqResp.Choices) > 0 {
		return groqResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from Groq")
}
