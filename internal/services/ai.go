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
	prompt := fmt.Sprintf(`Based on the following security scan results and auto-fix actions, provide a detailed report:

Scan Results & Actions:
%s

Please provide:
1. Executive Summary of Findings
2. Review of Auto-Fix Actions taken (if any)
3. Priority recommendations for remaining issues
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

// GenerateWorkflowJSON generates a workflow configuration from a prompt
func (s *AIService) GenerateWorkflowJSON(ctx context.Context, userPrompt string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert Workflow Builder Assistant.
Create a JSON configuration for a security workflow based on this request: "%s"

The JSON must return an object with "nodes" and "edges" arrays.
Node Types available: "trigger", "gobuster", "nikto", "nmap", "sqlmap", "wpscan", "owasp-vulnerabilities", "auto-fix", "email", "github-issue", "slack", "flow-chart".

Rules:
1. Always start with a "trigger" node.
2. Use logical "positions" (x, y) so nodes are laid out left-to-right (e.g. x: 0, x: 300, x: 600).
3. "edges" must connect nodes logically (source -> target).
4. Return ONLY valid JSON. No markdown formatting.

Example Structure:
{
  "nodes": [
    { "id": "1", "type": "trigger", "position": { "x": 0, "y": 100 }, "data": { "sourceUrl": "https://example.com" } },
    { "id": "2", "type": "nmap", "position": { "x": 300, "y": 100 }, "data": {} }
  ],
  "edges": [
    { "id": "e1-2", "source": "1", "target": "2" }
  ]
}`, userPrompt)

	if s.config.AI.GeminiAPIKey != "" {
		result, err := s.callGemini(ctx, prompt)
		if err == nil {
			// Clean markdown if present
			return cleanJSON(result), nil
		}
	}

	if s.config.AI.GroqAPIKey != "" {
		result, err := s.callGroq(ctx, prompt)
		if err == nil {
			return cleanJSON(result), nil
		}
	}

	return "", fmt.Errorf("no AI API keys configured")
}

func cleanJSON(s string) string {
	// Simple cleanup to remove ```json ... ``` wrapper if present
	if len(s) > 7 && s[:7] == "```json" {
		s = s[7:]
		if len(s) > 3 && s[len(s)-3:] == "```" {
			s = s[:len(s)-3]
		}
	}
	return s
}

// GenerateDocumentation generates project documentation using Groq
func (s *AIService) GenerateDocumentation(ctx context.Context, contextData string) (string, error) {
	prompt := fmt.Sprintf(`You are a Technical Writer. Generate comprehensive documentation for the following infrastructure and security context.
Return the response in Markdown format.

Context:
%s

Please generate:
1. A README.md content with:
   - Project Overview
   - Architecture Description
   - Security Posture (based on scan results)
   - Setup Instructions
2. An ARCHITECTURE.md content with:
   - Diagram description
   - Decision Records (ADRs) based on findings`, contextData)

	// User explicitly requested Groq
	if s.config.AI.GroqAPIKey != "" {
		return s.callGroq(ctx, prompt)
	}

	// Fallback to Gemini if Groq not configured
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
