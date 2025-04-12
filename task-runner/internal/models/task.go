package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskType represents the type of task
type TaskType string

const (
	TaskTypeHTTP TaskType = "http"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task represents a task to be executed
type Task struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Type        TaskType   `json:"type"`
	Schedule    string     `json:"schedule"` // cron expression
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   string     `json:"created_by"`
	Version     int        `json:"version"`
	Dependencies []uuid.UUID `json:"dependencies,omitempty"`
	Config      TaskConfig  `json:"config"`
}

// TaskConfig contains the configuration for different task types
type TaskConfig struct {
	HTTP *HTTPTaskConfig `json:"http,omitempty"`
}

// HTTPTaskConfig contains configuration for HTTP tasks
type HTTPTaskConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	ID        uuid.UUID  `json:"id"`
	TaskID    uuid.UUID  `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Output    string     `json:"output"`
	Error     string     `json:"error,omitempty"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Version   int        `json:"version"`
}

// NewTask creates a new task with default values
func NewTask(name string, taskType TaskType, schedule string, createdBy string) *Task {
	return &Task{
		ID:        uuid.New(),
		Name:      name,
		Type:      taskType,
		Schedule:  schedule,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: createdBy,
		Version:   1,
	}
} 