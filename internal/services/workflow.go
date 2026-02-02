package services

import (
	"fmt"

	"github.com/datmedevil17/go-vuln/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowService struct {
	db       *gorm.DB
	executor *WorkflowExecutor
}

func NewWorkflowService(db *gorm.DB, scannerService *ScannerService, notificationService *NotificationService, aiService *AIService, githubService *GitHubService) *WorkflowService {
	return &WorkflowService{
		db:       db,
		executor: NewWorkflowExecutor(db, scannerService, notificationService, aiService, githubService),
	}
}

// CreateWorkflow creates a new workflow
func (s *WorkflowService) CreateWorkflow(userID uuid.UUID, name string) (*models.Workflow, error) {
	workflow := &models.Workflow{
		UserID: userID,
		Name:   name,
		Nodes:  models.JSONArray{},
		Edges:  models.JSONArray{},
	}

	if err := s.db.Create(workflow).Error; err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowService) GetWorkflow(workflowID, userID uuid.UUID) (*models.Workflow, error) {
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

// ListWorkflows retrieves all workflows for a user
func (s *WorkflowService) ListWorkflows(userID uuid.UUID) ([]models.Workflow, error) {
	var workflows []models.Workflow
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, err
	}
	return workflows, nil
}

// UpdateWorkflow updates a workflow
func (s *WorkflowService) UpdateWorkflow(workflowID, userID uuid.UUID, updates map[string]interface{}) (*models.Workflow, error) {
	var workflow models.Workflow
	if err := s.db.Where("id = ? AND user_id = ?", workflowID, userID).First(&workflow).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&workflow).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return &workflow, nil
}

// DeleteWorkflow deletes a workflow
func (s *WorkflowService) DeleteWorkflow(workflowID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", workflowID, userID).Delete(&models.Workflow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ExecuteWorkflow executes a workflow asynchronously
func (s *WorkflowService) ExecuteWorkflow(workflow *models.Workflow, userID uuid.UUID) (*models.WorkflowExecution, error) {
	return s.executor.Execute(workflow, userID)
}

// ListWorkflowExecutions retrieves all workflow executions for a user with workflow names
func (s *WorkflowService) ListWorkflowExecutions(userID uuid.UUID) ([]models.WorkflowExecution, error) {
	var executions []models.WorkflowExecution

	// Use a join to get the workflow name
	err := s.db.Table("workflow_executions").
		Select("workflow_executions.*, workflows.name as name").
		Joins("left join workflows on workflows.id = workflow_executions.workflow_id").
		Where("workflow_executions.user_id = ?", userID).
		Order("workflow_executions.created_at DESC").
		Scan(&executions).Error

	if err != nil {
		return nil, err
	}

	// Calculate durations
	for i := range executions {
		if executions[i].StartedAt != nil && executions[i].CompletedAt != nil {
			executions[i].Duration = executions[i].CompletedAt.Sub(*executions[i].StartedAt).Milliseconds()
		}
	}

	return executions, nil
}

// GetExecution retrieves a specific workflow execution
func (s *WorkflowService) GetExecution(executionID, userID uuid.UUID) (*models.WorkflowExecution, error) {
	var execution models.WorkflowExecution
	if err := s.db.Where("id = ? AND user_id = ?", executionID, userID).First(&execution).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

// DeleteWorkflowExecution deletes a workflow execution report
func (s *WorkflowService) DeleteWorkflowExecution(executionID, userID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", executionID, userID).Delete(&models.WorkflowExecution{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
