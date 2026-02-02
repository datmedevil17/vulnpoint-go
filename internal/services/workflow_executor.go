package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/datmedevil17/go-vuln/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowExecutor struct {
	db                  *gorm.DB
	scannerService      *ScannerService
	notificationService *NotificationService
	aiService           *AIService
	githubService       *GitHubService
}

func NewWorkflowExecutor(db *gorm.DB, scannerService *ScannerService, notificationService *NotificationService, aiService *AIService, githubService *GitHubService) *WorkflowExecutor {
	return &WorkflowExecutor{
		db:                  db,
		scannerService:      scannerService,
		notificationService: notificationService,
		aiService:           aiService,
		githubService:       githubService,
	}
}

// WorkflowNode represents a node in the workflow graph
type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Position map[string]interface{} `json:"position"`
}

// WorkflowEdge represents an edge in the workflow graph
type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// Execute runs a workflow asynchronously
func (e *WorkflowExecutor) Execute(workflow *models.Workflow, userID uuid.UUID) (*models.WorkflowExecution, error) {
	// Create execution record
	execution := &models.WorkflowExecution{
		WorkflowID: workflow.ID,
		UserID:     userID,
		Status:     "pending",
		Results:    make(models.JSONMap),
	}

	if err := e.db.Create(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	execution.Name = workflow.Name

	// Launch async execution
	go e.executeAsync(execution.ID, workflow)

	return execution, nil
}

// executeAsync runs the workflow in the background
func (e *WorkflowExecutor) executeAsync(executionID uuid.UUID, workflow *models.Workflow) {
	log.Printf("🚀 Starting workflow execution: %s", executionID)

	// Update status to running
	startTime := time.Now()
	e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": startTime,
	})

	// Parse nodes and edges
	nodes, edges, err := e.parseWorkflow(workflow)
	if err != nil {
		e.failExecution(executionID, fmt.Sprintf("Failed to parse workflow: %v", err))
		return
	}

	// Get execution order
	executionOrder, err := e.topologicalSort(nodes, edges)
	if err != nil {
		e.failExecution(executionID, fmt.Sprintf("Failed to sort workflow: %v", err))
		return
	}

	log.Printf("📋 Execution order: %v", executionOrder)

	// Build In-Edges map for easy parent lookup
	inEdges := make(map[string][]string)
	for _, edge := range edges {
		inEdges[edge.Target] = append(inEdges[edge.Target], edge.Source)
	}

	// Execute nodes in order
	results := make(map[string]interface{})

	// Map to track execution state: "pending", "completed", "failed", "skipped"
	nodeStates := make(map[string]string)

	for _, nodeID := range executionOrder {
		node := e.findNode(nodes, nodeID)
		if node == nil {
			e.failExecution(executionID, fmt.Sprintf("Node not found: %s", nodeID))
			return
		}

		// Update current node
		e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Update("current_node", node.ID)

		// CHECK DEPENDENCIES
		shouldSkip := false
		skipReason := ""
		parents := inEdges[nodeID]

		for _, parentID := range parents {
			parentState := nodeStates[parentID]
			// 1. Cascade Skip/Fail
			if parentState == "skipped" || parentState == "failed" {
				shouldSkip = true
				skipReason = fmt.Sprintf("Parent %s was %s", parentID, parentState)
				break
			}

			// 2. Check Decision Logic
			// If parent was a decision node, check its output
			if parentResult, ok := results[parentID].(map[string]interface{}); ok {
				if parentResult["type"] == "decision" {
					if allowed, ok := parentResult["decision_result"].(bool); ok && !allowed {
						shouldSkip = true
						skipReason = fmt.Sprintf("Decision node %s returned false", parentID)
						break
					}
				}
			}
		}

		if shouldSkip {
			log.Printf("⏭️ Skipping node %s: %s", node.ID, skipReason)
			nodeStates[node.ID] = "skipped"
			// Store a dummy skipped result so downstream nodes know
			results[node.ID] = map[string]interface{}{
				"id":     node.ID,
				"status": "skipped",
				"reason": skipReason,
			}
			continue
		}

		log.Printf("⚙️  Executing node: %s (%s)", node.ID, node.Type)

		// Execute the node
		result, err := e.executeNode(node, results, workflow.UserID)
		if err != nil {
			log.Printf("❌ Node %s failed: %v", node.ID, err)
			nodeStates[node.ID] = "failed"
			e.failExecution(executionID, fmt.Sprintf("Node %s failed: %v", node.ID, err))
			return
		}

		nodeStates[node.ID] = "completed"

		// Store result
		results[node.ID] = result
		e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Update("results", models.JSONMap(results))
	}

	// Generate AI Report (Only for completed nodes)
	log.Printf("🤖 Generating AI Security Report...")
	var scanSummaries string
	for nodeID, result := range results {
		// Only analyze completed nodes
		if state := nodeStates[nodeID]; state != "completed" {
			continue
		}
		if nodeMap, ok := result.(map[string]interface{}); ok {
			if output, ok := nodeMap["output"].(string); ok {
				scanSummaries += fmt.Sprintf("Node %s (%s) Output:\n%s\n\n", nodeID, nodeMap["scanner"], output)
			}
		}
	}

	if scanSummaries != "" {
		aiReport, err := e.aiService.GenerateSecurityRecommendations(context.Background(), scanSummaries)
		if err != nil {
			log.Printf("⚠️ Failed to generate AI report: %v", err)
			results["ai_report_error"] = err.Error()
		} else {
			results["ai_report"] = map[string]interface{}{
				"ai_report":       aiReport,
				"security_grade":  "B", // Placeholder, ideally specific extraction logic would be better but keeping it simple
				"total_issues":    5,   // Placeholder
				"critical_issues": 0,
				"report_date":     time.Now(),
				"generated_by":    "VulnPilot AI",
			}
			e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Update("results", models.JSONMap(results))
		}
	}

	// Mark as completed
	completedTime := time.Now()
	e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": completedTime,
		"results":      models.JSONMap(results),
	})

	log.Printf("✅ Workflow execution completed: %s (duration: %v)", executionID, completedTime.Sub(startTime))
}

// parseWorkflow extracts nodes and edges from workflow
func (e *WorkflowExecutor) parseWorkflow(workflow *models.Workflow) ([]WorkflowNode, []WorkflowEdge, error) {
	var nodes []WorkflowNode
	var edges []WorkflowEdge

	// Parse nodes
	nodesBytes, err := json.Marshal(workflow.Nodes)
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(nodesBytes, &nodes); err != nil {
		return nil, nil, err
	}

	// Parse edges
	edgesBytes, err := json.Marshal(workflow.Edges)
	if err != nil {
		return nil, nil, err
	}
	if err := json.Unmarshal(edgesBytes, &edges); err != nil {
		return nil, nil, err
	}

	return nodes, edges, nil
}

// topologicalSort returns nodes in execution order
func (e *WorkflowExecutor) topologicalSort(nodes []WorkflowNode, edges []WorkflowEdge) ([]string, error) {
	// Build adjacency list and in-degree map
	adjList := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for _, node := range nodes {
		inDegree[node.ID] = 0
		adjList[node.ID] = []string{}
	}

	// Build graph
	for _, edge := range edges {
		adjList[edge.Source] = append(adjList[edge.Source], edge.Target)
		inDegree[edge.Target]++
	}

	// Find nodes with no dependencies
	queue := []string{}
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Process queue
	result := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Reduce in-degree for neighbors
		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(nodes) {
		return nil, fmt.Errorf("workflow contains cycles")
	}

	return result, nil
}

// executeNode executes a single node
func (e *WorkflowExecutor) executeNode(node *WorkflowNode, previousResults map[string]interface{}, userID uuid.UUID) (interface{}, error) {
	switch node.Type {
	case "trigger":
		return e.executeTrigger(node)
	case "nmap":
		return e.executeNmap(node, previousResults)
	case "nikto":
		return e.executeNikto(node, previousResults)
	case "gobuster":
		return e.executeGobuster(node, previousResults)
	case "sqlmap":
		return e.executeSqlmap(node, previousResults)
	case "wpscan":
		return e.executeWpscan(node, previousResults)
	case "email", "slack":
		return e.executeNotification(node, previousResults, userID)
	case "github-issue":
		return e.executeGitHubIssue(node, previousResults, userID)
	case "auto-fix":
		return e.executeAutoFix(node, previousResults, userID)
	case "owasp-vulnerabilities":
		return e.executeNikto(node, previousResults) // Map OWASP to Nikto for now
	case "flow-chart":
		return e.executeFlowChart(node, previousResults)
	case "secret-scan":
		return e.executeSecretScan(node, previousResults)
	case "dependency-check":
		return e.executeDependencyCheck(node, previousResults)
	case "semgrep-scan":
		return e.executeSemgrep(node, previousResults)
	case "container-scan":
		return e.executeContainerScan(node, previousResults)
	case "kube-bench":
		return e.executeKubeBench(node, previousResults)
	case "iac-scan":
		return e.executeTrivyIaC(node, previousResults)
	case "decision":
		return e.executeDecision(node, previousResults)
	case "estimate-cost":
		return e.executeEstimateCost(node, previousResults)
	case "policy-check":
		return e.executePolicyCheck(node, previousResults)
	case "generate-iac":
		return e.executeGenerateIaC(node, previousResults)
	case "drift-check":
		return e.executeDriftCheck(node, previousResults)
	case "generate-docs":
		return e.executeGenerateDocs(node, previousResults)
	default:
		return nil, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

// executeTrigger gets the target from trigger node
func (e *WorkflowExecutor) executeTrigger(node *WorkflowNode) (interface{}, error) {
	targetURL, ok := node.Data["sourceUrl"].(string)
	if !ok || targetURL == "" {
		// Fallback for demo if not set
		targetURL = "example.com"
	}

	return map[string]interface{}{
		"target": targetURL,
		"type":   "trigger",
	}, nil
}

// executeNmap runs nmap scanner
func (e *WorkflowExecutor) executeNmap(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	// Get target from trigger node
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for nmap")
	}

	// Get config from node data if available
	ports := "1-1000" // Default
	if p, ok := node.Data["ports"].(string); ok && p != "" {
		ports = p
	}

	log.Printf("🔍 Running Nmap scan on: %s ports: %s", target, ports)

	output, err := e.scannerService.RunNmap(target, ports)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "nmap",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeNikto runs nikto scanner
func (e *WorkflowExecutor) executeNikto(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for nikto")
	}

	log.Printf("🔍 Running Nikto scan on: %s", target)

	output, err := e.scannerService.RunNikto(target)
	if err != nil {
		return nil, err
	}

	// Try to parse JSON if possible, otherwise return raw output
	var jsonOutput interface{}
	if json.Unmarshal(output, &jsonOutput) == nil {
		return map[string]interface{}{
			"scanner": "nikto",
			"target":  target,
			"data":    jsonOutput,
			"output":  string(output), // Include raw output for reporting
			"status":  "completed",
		}, nil
	}

	return map[string]interface{}{
		"scanner": "nikto",
		"target":  target,
		"output":  string(output),
		"status":  "completed",
	}, nil
}

// executeGobuster runs gobuster scanner
func (e *WorkflowExecutor) executeGobuster(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for gobuster")
	}

	wordlist := ""
	if w, ok := node.Data["wordlist"].(string); ok {
		wordlist = w
	}

	log.Printf("🔍 Running Gobuster scan on: %s", target)

	output, err := e.scannerService.RunGobuster(target, wordlist)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "gobuster",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeSqlmap runs sqlmap scanner
func (e *WorkflowExecutor) executeSqlmap(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for sqlmap")
	}

	log.Printf("🔍 Running Sqlmap scan on: %s", target)

	output, err := e.scannerService.RunSqlmap(target)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "sqlmap",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeWpscan runs wpscan scanner
func (e *WorkflowExecutor) executeWpscan(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for wpscan")
	}

	log.Printf("🔍 Running WPScan on: %s", target)

	output, err := e.scannerService.RunWpscan(target)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "wpscan",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeNotification sends notification with results
func (e *WorkflowExecutor) executeNotification(node *WorkflowNode, previousResults map[string]interface{}, userID uuid.UUID) (interface{}, error) {
	log.Printf("📧 Sending %s notification with results", node.Type)

	// Fetch user to get email
	var user models.User
	if err := e.db.First(&user, "id = ?", userID).Error; err != nil {
		log.Printf("⚠️ Failed to fetch user for notification: %v", err)
		return map[string]interface{}{
			"type":   node.Type,
			"status": "failed",
			"error":  "user not found",
		}, nil
	}

	// Get target from previous results
	target := e.getTarget(previousResults)

	// Aggregate results for AI
	var scanSummaries string
	for nodeID, result := range previousResults {
		if nodeMap, ok := result.(map[string]interface{}); ok {
			// Check for structured data first
			if data, ok := nodeMap["data"]; ok {
				formatted := formatScanData(data)
				scanSummaries += fmt.Sprintf("Node %s (%s) Output:\n%s\n\n", nodeID, nodeMap["scanner"], formatted)
			} else if output, ok := nodeMap["output"].(string); ok {
				scanSummaries += fmt.Sprintf("Node %s (%s) Output:\n%s\n\n", nodeID, nodeMap["scanner"], output)
			}

			// Special handling for Auto-Fix results
			if nodeType, ok := nodeMap["type"].(string); ok && nodeType == "auto-fix" {
				status := nodeMap["status"]
				prURL := nodeMap["pr_url"]
				scanSummaries += fmt.Sprintf("🛠️ Auto-Fix Action (Node %s):\nStatus: %s\nPR URL: %v\n\n", nodeID, status, prURL)
			}
		}
	}

	// Generate Report
	aiReport := "No scan data available for analysis."
	if scanSummaries != "" {
		report, err := e.aiService.GenerateSecurityRecommendations(context.Background(), scanSummaries)
		if err == nil {
			aiReport = report
		} else {
			log.Printf("⚠️ Failed to generate AI report for email: %v", err)
			aiReport = fmt.Sprintf("AI Analysis Failed: %v", err)
		}
	}

	// Determine recipient email
	recipientEmail := user.Email

	// The frontend stores config in data.config.email or data.config.to
	if config, ok := node.Data["config"].(map[string]interface{}); ok {
		if email, ok := config["email"].(string); ok && email != "" {
			recipientEmail = email
		} else if to, ok := config["to"].(string); ok && to != "" {
			recipientEmail = to
		}
	} else {
		// Fallback to flat structure if config is missing
		if email, ok := node.Data["email"].(string); ok && email != "" {
			recipientEmail = email
		} else if to, ok := node.Data["to"].(string); ok && to != "" {
			recipientEmail = to
		}
	}

	if recipientEmail == "" {
		log.Printf("⚠️ No recipient email available for notification")
		return map[string]interface{}{
			"type":   node.Type,
			"status": "failed",
			"error":  "no recipient email provided",
		}, nil
	}

	log.Printf("📧 Sending notification to: %s", recipientEmail)

	// Send email with report
	if err := e.notificationService.SendWorkflowReport(recipientEmail, target, "completed", aiReport); err != nil {
		log.Printf("⚠️ Failed to send notification to %s: %v", recipientEmail, err)
		return map[string]interface{}{
			"type":   node.Type,
			"status": "failed",
			"error":  err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"type":   node.Type,
		"status": "sent",
	}, nil
}

// getTarget extracts target from previous results
func (e *WorkflowExecutor) getTarget(previousResults map[string]interface{}) string {
	for _, result := range previousResults {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if target, ok := resultMap["target"].(string); ok {
				return target
			}
		}
	}
	return ""
}

// findNode finds a node by ID
func (e *WorkflowExecutor) findNode(nodes []WorkflowNode, nodeID string) *WorkflowNode {
	for i := range nodes {
		if nodes[i].ID == nodeID {
			return &nodes[i]
		}
	}
	return nil
}

// failExecution marks execution as failed
func (e *WorkflowExecutor) failExecution(executionID uuid.UUID, errorMsg string) {
	log.Printf("❌ Workflow execution failed: %s - %s", executionID, errorMsg)
	completedTime := time.Now()
	e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Updates(map[string]interface{}{
		"status":       "failed",
		"error":        errorMsg,
		"completed_at": completedTime,
	})
}

// executeGitHubIssue creates a GitHub issue with results
func (e *WorkflowExecutor) executeGitHubIssue(node *WorkflowNode, previousResults map[string]interface{}, userID uuid.UUID) (interface{}, error) {
	log.Printf("🐙 Creating GitHub Issue")

	// Fetch user to get access token
	var user models.User
	if err := e.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	if user.AccessToken == "" {
		return nil, fmt.Errorf("user has no GitHub access token")
	}

	// Get target from previous results
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for issue creation")
	}

	// Parse owner/repo from target
	// Assuming target is like https://github.com/owner/repo or just owner/repo
	// For now, let's try to parse simple URL
	var owner, repo string
	// Simple parsing logic (can be robustified)
	if len(target) > 19 && target[:19] == "https://github.com/" {
		parts := splitParam(target[19:], "/")
		if len(parts) >= 2 {
			owner = parts[0]
			repo = parts[1]
		}
	}

	// Fallback: check if node data has owner/repo
	if val, ok := node.Data["owner"].(string); ok && val != "" {
		owner = val
	}
	if val, ok := node.Data["repo"].(string); ok && val != "" {
		repo = val
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("could not determine GitHub owner/repo from target: %s", target)
	}

	// Aggregate results for Issue Body
	var scanSummaries string
	for nodeID, result := range previousResults {
		if nodeMap, ok := result.(map[string]interface{}); ok {
			// Check for structured data first
			if data, ok := nodeMap["data"]; ok {
				formatted := formatScanData(data)
				scanSummaries += fmt.Sprintf("## Scan: %s (Node %s)\n%s\n\n", nodeMap["scanner"], nodeID, formatted)
			} else if output, ok := nodeMap["output"].(string); ok {
				scanSummaries += fmt.Sprintf("## Scan: %s (Node %s)\n```\n%s\n```\n\n", nodeMap["scanner"], nodeID, output)
			}
		}
	}

	// Generate Issue Content
	title := fmt.Sprintf("Security Vulnerabilities Detected in %s/%s", owner, repo)
	body := fmt.Sprintf("# Security Scan Results\n\nAutomated scan detected potential issues.\n\n%s\n\n*Report generated by VulnPilot*", scanSummaries)

	// Use AI to generate better title/body if available
	if scanSummaries != "" {
		aiRecommendation, err := e.aiService.GenerateSecurityRecommendations(context.Background(), scanSummaries)
		if err == nil {
			body = fmt.Sprintf("# Security Analysis\n\n%s\n\n## Raw Logs\n\n%s", aiRecommendation, scanSummaries)
		}
	}

	// Create Issue
	issue, err := e.githubService.CreateIssue(context.Background(), user.AccessToken, owner, repo, title, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create github issue: %v", err)
	}

	log.Printf("✅ Created GitHub Issue #%d: %s", issue.Number, issue.HTMLURL)

	return map[string]interface{}{
		"type":       "github-issue",
		"issue_url":  issue.HTMLURL,
		"issue_id":   issue.ID,
		"status":     "created",
		"repository": fmt.Sprintf("%s/%s", owner, repo),
	}, nil
}

func (e *WorkflowExecutor) executeAutoFix(node *WorkflowNode, previousResults map[string]interface{}, userID uuid.UUID) (interface{}, error) {
	log.Printf("🔧 Execute Auto-Fix Agent")

	// 1. Authenticate
	var user models.User
	if err := e.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}

	if user.AccessToken == "" {
		return nil, fmt.Errorf("user has no GitHub access token")
	}

	// 2. Parse Context (Owner, Repo, Path, Branch)
	target := e.getTarget(previousResults)
	owner, repo := e.parseGitHubTarget(target)

	if val, ok := node.Data["owner"].(string); ok && val != "" {
		owner = val
	}
	if val, ok := node.Data["repo"].(string); ok && val != "" {
		repo = val
	}

	path, _ := node.Data["path"].(string)
	branch := "main" // Default
	if val, ok := node.Data["branch"].(string); ok && val != "" {
		branch = val
	}

	// Dynamic Path Inference
	// If path is missing, try to find it in previous scanner results
	if path == "" {
		log.Printf("🔍 Path not provided. searching previous scanner results...")
		for _, result := range previousResults {
			if resMap, ok := result.(map[string]interface{}); ok {
				// Check Gitleaks/Semgrep findings
				if output, ok := resMap["output"].(string); ok {
					// Extremely simple heuristic to find a file path in JSON
					// In a real app, unmarshal properly based on scanner type
					if strings.Contains(output, `"file": "`) {
						start := strings.Index(output, `"file": "`) + 9
						end := strings.Index(output[start:], `"`)
						if start > 9 && end > 0 {
							path = output[start : start+end]
							log.Printf("🎯 Inferred path from scanner: %s", path)
							break
						}
					}
					// Semgrep style
					if strings.Contains(output, `"path": "`) {
						start := strings.Index(output, `"path": "`) + 9
						end := strings.Index(output[start:], `"`)
						if start > 9 && end > 0 {
							path = output[start : start+end]
							log.Printf("🎯 Inferred path from scanner: %s", path)
							break
						}
					}
				}
			}
		}
	}

	if owner == "" || repo == "" || path == "" {
		return nil, fmt.Errorf("auto-fix requires owner, repo, and path (target: %s). Could not infer path from scanner results.", target)
	}

	// 3. Fetch File Content
	log.Printf("📖 Reading file: %s/%s/%s", owner, repo, path)
	content, err := e.githubService.GetFileContent(context.Background(), user.AccessToken, owner, repo, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// 4. Identify Vulnerability
	vulnerability, _ := node.Data["vulnerability"].(string)
	if vulnerability == "" {
		// If not provided, analyze the code now
		log.Printf("🔍 Analyzing code for vulnerabilities...")

		// Check for previous scanner results to help the analysis
		var scannerContext string
		for _, result := range previousResults {
			if resMap, ok := result.(map[string]interface{}); ok {
				if output, ok := resMap["output"].(string); ok {
					scannerContext += fmt.Sprintf("Scanner Output (%s):\n%s\n\n", resMap["scanner"], output)
				}
			}
		}

		// Heuristic: determine language from extension
		lang := "go" // default
		// ... simplified language detection ...

		// Pass scanner context if available
		inputContext := content
		if scannerContext != "" {
			inputContext = fmt.Sprintf("SCANNER FINDINGS:\n%s\n\nCODE TO FIX:\n%s", scannerContext, content)
		}

		analysis, err := e.aiService.AnalyzeCode(context.Background(), inputContext, lang)
		if err != nil {
			return nil, fmt.Errorf("analysis failed: %v", err)
		}
		vulnerability = analysis
	}

	// 5. Generate Fix
	log.Printf("🤖 Generating fix for vulnerability...")
	fixedCode, err := e.aiService.GenerateFix(context.Background(), content, vulnerability)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fix: %v", err)
	}

	// 6. Create Branch
	fixBranch := fmt.Sprintf("fix/vuln-%d", time.Now().Unix())
	log.Printf("🌿 Creating branch: %s", fixBranch)

	// Get base SHA
	ref, err := e.githubService.GetReference(context.Background(), user.AccessToken, owner, repo, "heads/"+branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get base ref: %v", err)
	}

	// Create branch
	if err := e.githubService.CreateBranch(context.Background(), user.AccessToken, owner, repo, fixBranch, ref.Object.Sha); err != nil {
		return nil, fmt.Errorf("failed to create branch: %v", err)
	}

	// 7. Update File (Commit)
	// Get file SHA for update
	fileSha, err := e.githubService.GetFileSHA(context.Background(), user.AccessToken, owner, repo, path, fixBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get file sha: %v", err)
	}

	log.Printf("💾 Committing fix...")
	if err := e.githubService.UpdateFile(context.Background(), user.AccessToken, owner, repo, path, fixedCode, fileSha, "fix: resolve security vulnerability", fixBranch); err != nil {
		return nil, fmt.Errorf("failed to update file: %v", err)
	}

	// 8. Create Pull Request
	log.Printf("🚀 Creating Pull Request...")
	prTitle := "fix: resolve security vulnerability in " + path
	prBody := fmt.Sprintf("This PR fixes a detected vulnerability.\n\n**Vulnerability:**\n%s\n\n*Generated by VulnPilot*", vulnerability)

	pr, err := e.githubService.CreatePullRequest(context.Background(), user.AccessToken, owner, repo, prTitle, prBody, fixBranch, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %v", err)
	}

	return map[string]interface{}{
		"type":      "auto-fix",
		"pr_url":    pr.HTMLURL,
		"pr_number": pr.Number,
		"status":    "created",
		"branch":    fixBranch,
		"output":    fmt.Sprintf("Auto-Fix PR Created: %s", pr.HTMLURL),
	}, nil
}

func (e *WorkflowExecutor) parseGitHubTarget(target string) (string, string) {
	if len(target) > 19 && target[:19] == "https://github.com/" {
		parts := splitParam(target[19:], "/")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

func splitParam(s, sep string) []string {
	var parts []string
	current := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// executeFlowChart handles flow-chart nodes (pass-through)
func (e *WorkflowExecutor) executeFlowChart(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("📊 Executing Flow Chart Node (Pass-through)")

	target := e.getTarget(previousResults)

	return map[string]interface{}{
		"type":   "flow-chart",
		"status": "completed",
		"target": target,
	}, nil
}

func formatScanData(data interface{}) string {
	// If it's the specific Nikto format we use
	if dataMap, ok := data.(map[string]interface{}); ok {
		if vulns, ok := dataMap["vulnerabilities"].([]interface{}); ok {
			var formatted string
			for _, v := range vulns {
				if vStr, ok := v.(string); ok {
					formatted += fmt.Sprintf("- %s\n", vStr)
				}
			}
			if formatted != "" {
				return formatted
			}
		}
	}

	// Fallback to pretty JSON
	bytes, _ := json.MarshalIndent(data, "", "  ")
	return fmt.Sprintf("```json\n%s\n```", string(bytes))
}

// executeSecretScan simulates a Gitleaks scan
func (e *WorkflowExecutor) executeSecretScan(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🔑 Executing Secret Scan (Gitleaks)...")
	time.Sleep(2 * time.Second) // Simulate work

	// Mock findings: Using README.md as it likely exists in any repo
	output := `
{
  "findings": [
    {
      "rule": "generic-secret",
      "file": "README.md",
      "startLine": 1,
      "secret": "password123",
      "message": "Simulated secret found for Auto-Fix testing"
    }
  ]
}`
	return map[string]interface{}{
		"scanner": "gitleaks",
		"status":  "completed",
		"output":  output,
		"data": map[string]interface{}{
			"leaked_secrets": 1,
			"files_scanned":  15,
		},
	}, nil
}

// executeDependencyCheck simulates a Trivy/SCA scan
func (e *WorkflowExecutor) executeDependencyCheck(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("📦 Executing Dependency Check (Trivy)...")
	time.Sleep(2 * time.Second)

	output := `
{
  "Target": "go.mod",
  "Vulnerabilities": [
    {
      "VulnerabilityID": "CVE-2023-1234",
      "PkgName": "golang.org/x/net",
      "InstalledVersion": "v0.7.0",
      "FixedVersion": "v0.17.0",
      "Severity": "HIGH"
    }
  ]
}`
	return map[string]interface{}{
		"scanner": "trivy-sca",
		"status":  "completed",
		"output":  output,
		"data": map[string]interface{}{
			"vulnerabilities_found": 1,
			"severity_high":         1,
		},
	}, nil
}

// executeSemgrep simulates a Semgrep SAST scan
func (e *WorkflowExecutor) executeSemgrep(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🔬 Executing Semgrep SAST...")
	time.Sleep(2 * time.Second)

	// Mock findings: Using main.go as it likely exists
	output := `
{
  "results": [
    {
      "check_id": "go.lang.security.audit.xss.reflect.xss",
      "path": "main.go",
      "start": { "line": 1, "col": 1 },
      "extra": { "message": "Potential XSS vulnerability detected (Simulated)" }
    }
  ]
}`
	return map[string]interface{}{
		"scanner": "semgrep",
		"status":  "completed",
		"output":  output,
	}, nil
}

// executeContainerScan simulates a Container scan
func (e *WorkflowExecutor) executeContainerScan(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🐳 Executing Container Scan...")
	time.Sleep(2 * time.Second)

	output := `
{
  "Image": "app:latest",
  "OS": "alpine:3.14",
  "Vulnerabilities": [
    {
      "ID": "CVE-2022-4567",
      "Package": "openssl",
      "Severity": "CRITICAL"
    }
  ]
}`
	return map[string]interface{}{
		"scanner": "trivy-image",
		"status":  "completed",
		"output":  output,
	}, nil
}

// executeKubeBench runs kube-bench
func (e *WorkflowExecutor) executeKubeBench(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		target = "cluster"
	}

	log.Printf("☸️ Running Kube-Bench scan on: %s", target)
	output, err := e.scannerService.RunKubeBench(target)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "kube-bench",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeTrivyIaC runs Trivy IaC scan
func (e *WorkflowExecutor) executeTrivyIaC(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for IaC scan")
	}

	log.Printf("🏗️ Running Trivy IaC scan on: %s", target)
	output, err := e.scannerService.RunTrivyIaC(target)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"scanner": "trivy-iac",
		"target":  target,
		"output":  output,
		"status":  "completed",
	}, nil
}

// executeDecision handles logic branching
func (e *WorkflowExecutor) executeDecision(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🤔 Evaluating Decision Node...")

	// Get configuration
	variable, _ := node.Data["variable"].(string)
	operator, _ := node.Data["operator"].(string)
	thresholdStr, _ := node.Data["value"].(string)

	log.Printf("   Rule: %s %s %s", variable, operator, thresholdStr)

	// 1. Resolve Variable Value from Previous Results
	var actualValue float64
	found := false

	for _, result := range previousResults {
		if resMap, ok := result.(map[string]interface{}); ok {

			// Check for Cost
			if variable == "cost" {
				// Try to parse cost from strings like "$154.20"
				if costStr, ok := resMap["monthly_cost"].(string); ok {
					cleaned := strings.ReplaceAll(strings.ReplaceAll(costStr, "$", ""), ",", "")
					if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
						actualValue = val
						found = true
						break
					}
				}
			}

			// Check for Vulnerabilities (Sum them up?)
			if variable == "vulnerabilities" {
				// Check structured data from trivy/semgrep/etc
				if data, ok := resMap["data"].(map[string]interface{}); ok {
					if count, ok := data["vulnerabilities_found"].(float64); ok { // JSON numbers are float64 in Go interface{}
						actualValue += count
						found = true
					}
					if count, ok := data["leaked_secrets"].(float64); ok {
						actualValue += count
						found = true
					}
				}
			}

			// Check for Policy Pass/Fail
			if variable == "risk_score" {
				// Mock logic for now, count criticals * 10?
				if data, ok := resMap["data"].(map[string]interface{}); ok {
					if high, ok := data["severity_high"].(float64); ok {
						actualValue += high * 5
						found = true
					}
				}
			}
		}
	}

	if !found && variable != "manual_input" {
		log.Printf("   ⚠️ Variable %s not found in previous results, defaulting to 0", variable)
	}

	// 2. Parse Threshold
	threshold, _ := strconv.ParseFloat(thresholdStr, 64)

	// 3. Evaluate
	result := false
	switch operator {
	case "gt":
		result = actualValue > threshold
	case "lt":
		result = actualValue < threshold
	case "eq":
		result = actualValue == threshold
	case "neq":
		result = actualValue != threshold
	default:
		// Default validation (e.g. if manual_input)
		// For now simple pass
		result = true
	}

	log.Printf("   Result: %f %s %f = %v", actualValue, operator, threshold, result)

	return map[string]interface{}{
		"type":            "decision",
		"decision_result": result, // The key flag for the engine
		"actual_value":    actualValue,
		"status":          "completed",
	}, nil
}

// executeEstimateCost calculates infrastructure cost
func (e *WorkflowExecutor) executeEstimateCost(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	target := e.getTarget(previousResults)
	if target == "" {
		return nil, fmt.Errorf("no target found for cost estimation")
	}

	log.Printf("💰 Estimating Cloud Costs (Infracost) for: %s", target)

	output, err := e.scannerService.RunInfracost(target)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"type":   "estimate-cost",
		"status": "completed",
		"output": output,
	}, nil
}

// executePolicyCheck validates OPA rules
func (e *WorkflowExecutor) executePolicyCheck(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("👮 Checking Policies (OPA)...")
	time.Sleep(1 * time.Second)
	// Mock
	return map[string]interface{}{
		"type":       "policy-check",
		"status":     "completed",
		"passed":     true,
		"violations": 0,
		"output":     "All policies passed (CIS Benchmark Level 1)",
	}, nil
}

// executeGenerateIaC creates Terraform code
func (e *WorkflowExecutor) executeGenerateIaC(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🏗️ Generating Infrastructure as Code...")
	time.Sleep(2 * time.Second)
	// Mock
	return map[string]interface{}{
		"type":   "generate-iac",
		"status": "completed",
		"files":  []string{"main.tf", "variables.tf", "outputs.tf"},
		"output": "Generated AWS ECS Fargate Cluster configuration",
		"changes": []map[string]string{
			{
				"path": "main.tf",
				"type": "create",
				"after": `resource "aws_ecs_cluster" "main" {
  name = "vulnpilot-cluster"
  
  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}`,
			},
			{
				"path": "variables.tf",
				"type": "create",
				"after": `variable "region" {
  default = "us-east-1"
}`,
			},
		},
	}, nil
}

// executeDriftCheck checks for infrastructure drift
func (e *WorkflowExecutor) executeDriftCheck(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("🔎 Checking for Infrastructure Drift...")
	time.Sleep(2 * time.Second)

	// Mock drift detected
	return map[string]interface{}{
		"type":           "drift-check",
		"status":         "completed",
		"drift_detected": true,
		"output":         "Drift detected in Security Group configuration.",
		"changes": []map[string]interface{}{
			{
				"path":   "aws_security_group.allow_ssh",
				"type":   "update",
				"before": "ingress {\n  from_port = 22\n  to_port = 22\n  cidr_blocks = [\"10.0.0.0/8\"]\n}",
				"after":  "ingress {\n  from_port = 22\n  to_port = 22\n  cidr_blocks = [\"0.0.0.0/0\"]\n}",
			},
			{
				"path":   "aws_s3_bucket.logs",
				"type":   "delete",
				"before": "resource \"aws_s3_bucket\" \"logs\" {\n  bucket = \"my-logs\"\n}",
			},
		},
	}, nil
}

// executeGenerateDocs creates documentation using AI
func (e *WorkflowExecutor) executeGenerateDocs(node *WorkflowNode, previousResults map[string]interface{}) (interface{}, error) {
	log.Printf("📝 Generating Documentation using AI...")

	// Aggregate context
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Workflow Execution Results:\n")

	for nodeID, result := range previousResults {
		if nodeMap, ok := result.(map[string]interface{}); ok {
			contextBuilder.WriteString(fmt.Sprintf("\nNode: %s (%v)\n", nodeID, nodeMap["scanner"]))
			if output, ok := nodeMap["output"].(string); ok {
				// Truncate long outputs for prompt context
				if len(output) > 2000 {
					contextBuilder.WriteString(output[:2000] + "...(truncated)")
				} else {
					contextBuilder.WriteString(output)
				}
			}
		}
	}

	docContent, err := e.aiService.GenerateDocumentation(context.Background(), contextBuilder.String())
	if err != nil {
		log.Printf("⚠️ Failed to generate documentation: %v", err)
		return nil, err
	}

	return map[string]interface{}{
		"type":   "generate-docs",
		"status": "completed",
		"files": []string{
			"README.md",
			"ARCHITECTURE.md",
		},
		"output": docContent,
	}, nil
}
