package handlers

import (
	"log"

	"github.com/datmedevil17/go-vuln/internal/middleware"
	"github.com/datmedevil17/go-vuln/internal/models"
	"github.com/datmedevil17/go-vuln/internal/services"
	"github.com/datmedevil17/go-vuln/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowHandler struct {
	workflowService *services.WorkflowService
}

type CreateWorkflowRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateWorkflowRequest struct {
	Name            *string        `json:"name,omitempty"`
	Nodes           *[]interface{} `json:"nodes,omitempty"`
	Edges           *[]interface{} `json:"edges,omitempty"`
	IsActive        *bool          `json:"is_active,omitempty"`
	ScheduleEnabled *bool          `json:"schedule_enabled,omitempty"`
	ScheduleFreq    *string        `json:"schedule_frequency,omitempty"`
}

func NewWorkflowHandler(workflowService *services.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		workflowService: workflowService,
	}
}

// CreateWorkflow creates a new workflow
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request: "+err.Error())
		return
	}

	workflow, err := h.workflowService.CreateWorkflow(userID, req.Name)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to create workflow")
		return
	}

	utils.SuccessMessageResponse(c, "Workflow created successfully", workflow)
}

// GetWorkflow retrieves a specific workflow
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid workflow ID")
		return
	}

	workflow, err := h.workflowService.GetWorkflow(workflowID, userID)
	if err != nil {
		utils.NotFoundResponse(c, "Workflow not found")
		return
	}

	utils.SuccessResponse(c, workflow)
}

// ListWorkflows retrieves all workflows for the user
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	workflows, err := h.workflowService.ListWorkflows(userID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch workflows")
		return
	}

	utils.SuccessResponse(c, workflows)
}

// UpdateWorkflow updates a workflow
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid workflow ID")
		return
	}

	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request: "+err.Error())
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Nodes != nil {
		updates["nodes"] = models.JSONArray(*req.Nodes)
	}
	if req.Edges != nil {
		updates["edges"] = models.JSONArray(*req.Edges)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ScheduleEnabled != nil {
		updates["schedule_enabled"] = *req.ScheduleEnabled
	}
	if req.ScheduleFreq != nil {
		updates["schedule_frequency"] = *req.ScheduleFreq
	}

	workflow, err := h.workflowService.UpdateWorkflow(workflowID, userID, updates)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Workflow not found")
			return
		}
		log.Printf("Error updating workflow %s: %v", workflowID, err)
		utils.InternalErrorResponse(c, "Failed to update workflow: "+err.Error())
		return
	}

	utils.SuccessMessageResponse(c, "Workflow updated successfully", workflow)
}

// DeleteWorkflow deletes a workflow
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid workflow ID")
		return
	}

	if err := h.workflowService.DeleteWorkflow(workflowID, userID); err != nil {
		utils.InternalErrorResponse(c, "Failed to delete workflow")
		return
	}

	utils.SuccessMessageResponse(c, "Workflow deleted successfully", nil)
}

// ExecuteWorkflow executes a workflow
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid workflow ID")
		return
	}

	// Get workflow
	workflow, err := h.workflowService.GetWorkflow(workflowID, userID)
	if err != nil {
		utils.NotFoundResponse(c, "Workflow not found")
		return
	}

	// Execute workflow asynchronously
	execution, err := h.workflowService.ExecuteWorkflow(workflow, userID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to start execution")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message":      "Workflow execution started",
		"execution_id": execution.ID.String(),
		"workflow_id":  workflowID.String(),
	})
}

// ListWorkflowExecutions retrieves all workflow executions
func (h *WorkflowHandler) ListWorkflowExecutions(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	executions, err := h.workflowService.ListWorkflowExecutions(userID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to fetch workflow executions")
		return
	}

	utils.SuccessResponse(c, executions)
}

// GetExecution retrieves a specific workflow execution
func (h *WorkflowHandler) GetExecution(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid execution ID")
		return
	}

	execution, err := h.workflowService.GetExecution(executionID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Execution not found")
			return
		}
		utils.InternalErrorResponse(c, "Failed to fetch execution")
		return
	}

	utils.SuccessResponse(c, execution)
}

// DeleteWorkflowExecution deletes a workflow execution
func (h *WorkflowHandler) DeleteWorkflowExecution(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	executionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequestResponse(c, "Invalid execution ID")
		return
	}

	if err := h.workflowService.DeleteWorkflowExecution(executionID, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.NotFoundResponse(c, "Report not found")
			return
		}
		utils.InternalErrorResponse(c, "Failed to delete report")
		return
	}

	utils.SuccessMessageResponse(c, "Report deleted successfully", nil)
}
