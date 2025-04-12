package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/task-runner/internal/models"
	"github.com/yourusername/task-runner/internal/storage"
)

// Worker handles task execution
type Worker struct {
	storage storage.Storage
	client  *http.Client
}

// NewWorker creates a new worker instance
func NewWorker(storage storage.Storage) *Worker {
	return &Worker{
		storage: storage,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteTask executes a task and stores its result
func (w *Worker) ExecuteTask(ctx context.Context, task *models.Task) (*models.TaskResult, error) {
	result := &models.TaskResult{
		ID:        uuid.New(),
		TaskID:    task.ID,
		Status:    models.TaskStatusRunning,
		StartTime: time.Now(),
		Version:   task.Version,
	}

	// Create a channel to handle task completion
	done := make(chan struct{})
	var output string
	var execErr error

	// Execute task in a goroutine
	go func() {
		defer close(done)
		switch task.Type {
		case models.TaskTypeHTTP:
			output, execErr = w.executeHTTPTask(ctx, task)
		default:
			execErr = fmt.Errorf("unsupported task type: %s", task.Type)
		}
	}()

	// Wait for task completion or context cancellation
	select {
	case <-done:
		// Update result based on execution
		result.EndTime = time.Now()
		if execErr != nil {
			result.Status = models.TaskStatusFailed
			result.Error = execErr.Error()
		} else {
			result.Status = models.TaskStatusCompleted
			result.Output = output
		}
	case <-ctx.Done():
		// Context was cancelled
		result.EndTime = time.Now()
		result.Status = models.TaskStatusFailed
		result.Error = ctx.Err().Error()
	}

	return result, nil
}

// executeHTTPTask executes an HTTP task
func (w *Worker) executeHTTPTask(ctx context.Context, task *models.Task) (string, error) {
	if task.Config.HTTP == nil {
		return "", fmt.Errorf("HTTP config is required for HTTP tasks")
	}

	config := task.Config.HTTP
	req, err := http.NewRequestWithContext(ctx, config.Method, config.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range config.Headers {
		req.Header.Add(key, value)
	}

	// Execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	output, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(output), nil
} 