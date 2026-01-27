package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
}

func NewWorkflowExecutor(db *gorm.DB, scannerService *ScannerService, notificationService *NotificationService, aiService *AIService) *WorkflowExecutor {
	return &WorkflowExecutor{
		db:                  db,
		scannerService:      scannerService,
		notificationService: notificationService,
		aiService:           aiService,
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

	// Execute nodes in order
	results := make(map[string]interface{})
	for _, nodeID := range executionOrder {
		node := e.findNode(nodes, nodeID)
		if node == nil {
			e.failExecution(executionID, fmt.Sprintf("Node not found: %s", nodeID))
			return
		}

		// Update current node
		e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Update("current_node", node.Type)

		log.Printf("⚙️  Executing node: %s (%s)", node.ID, node.Type)

		// Execute the node
		result, err := e.executeNode(node, results, workflow.UserID)
		if err != nil {
			e.failExecution(executionID, fmt.Sprintf("Node %s failed: %v", node.ID, err))
			return
		}

		// Store result
		results[node.ID] = result
		e.db.Model(&models.WorkflowExecution{}).Where("id = ?", executionID).Update("results", models.JSONMap(results))
	}

	// Generate AI Report
	log.Printf("🤖 Generating AI Security Report...")
	var scanSummaries string
	for nodeID, result := range results {
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
	case "email", "github-issue", "slack":
		return e.executeNotification(node, previousResults, userID)
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
			if output, ok := nodeMap["output"].(string); ok {
				scanSummaries += fmt.Sprintf("Node %s (%s) Output:\n%s\n\n", nodeID, nodeMap["scanner"], output)
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
