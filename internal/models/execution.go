package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowExecution struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	WorkflowID  uuid.UUID  `gorm:"type:uuid;not null" json:"workflowId"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	Status      string     `gorm:"default:'pending'" json:"status"` // pending, running, completed, failed
	CurrentNode string     `json:"currentNode,omitempty"`
	Results     JSONMap    `gorm:"type:jsonb;default:'{}'" json:"results"` // Node ID -> Result
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Name        string     `gorm:"->" json:"name"`    // Workflow name, joined from workflows table
	Duration    int64      `gorm:"-" json:"duration"` // Duration in milliseconds
}

// JSONMap custom type for handling JSONB maps
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

func (w *WorkflowExecution) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
