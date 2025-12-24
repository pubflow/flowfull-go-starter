package models

import (
	"time"

	"gorm.io/gorm"
)

// Task represents a task in the system (example model)
type Task struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	UserID      string         `gorm:"index;not null" json:"user_id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Status      string         `gorm:"default:'pending'" json:"status"` // pending, in_progress, completed
	Priority    string         `gorm:"default:'medium'" json:"priority"` // low, medium, high
	DueDate     *time.Time     `json:"due_date,omitempty"`
}

// TableName specifies the table name for Task model
func (Task) TableName() string {
	return "tasks"
}

