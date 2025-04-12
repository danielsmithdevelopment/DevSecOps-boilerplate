package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/yourusername/task-runner/internal/models"
	"github.com/yourusername/task-runner/internal/storage"
	"github.com/yourusername/task-runner/internal/worker"
)

// Scheduler manages task scheduling and execution
type Scheduler struct {
	storage  storage.Storage
	worker   *worker.Worker
	cron     *cron.Cron
	mu       sync.RWMutex
	entries  map[uuid.UUID]cron.EntryID
	stopChan chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(storage storage.Storage, worker *worker.Worker) *Scheduler {
	return &Scheduler{
		storage:  storage,
		worker:   worker,
		cron:     cron.New(),
		entries:  make(map[uuid.UUID]cron.EntryID),
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Load existing tasks
	tasks, err := s.storage.ListTasks(ctx)
	if err != nil {
		return err
	}

	// Schedule existing tasks
	for _, task := range tasks {
		if err := s.ScheduleTask(ctx, task); err != nil {
			return err
		}
	}

	s.cron.Start()
	return nil
}

// Stop halts the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	close(s.stopChan)
}

// ScheduleTask schedules a task for execution
func (s *Scheduler) ScheduleTask(ctx context.Context, task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing schedule if any
	if entryID, exists := s.entries[task.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, task.ID)
	}

	// Create a copy of the task to avoid race conditions
	taskCopy := *task

	// Schedule the task
	entryID, err := s.cron.AddFunc(task.Schedule, func() {
		// Create a new context for each execution
		execCtx := context.Background()
		s.executeTask(execCtx, &taskCopy)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule task %s with schedule %s: %w", task.ID, task.Schedule, err)
	}

	s.entries[task.ID] = entryID
	return nil
}

// executeTask executes a task and handles dependencies
func (s *Scheduler) executeTask(ctx context.Context, task *models.Task) {
	fmt.Printf("Starting execution of task %s\n", task.ID)

	// Create a new context with timeout for this task execution
	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check dependencies
	if len(task.Dependencies) > 0 {
		fmt.Printf("Checking dependencies for task %s\n", task.ID)
		for _, depID := range task.Dependencies {
			// Check if dependency is completed
			results, err := s.storage.ListTaskResults(taskCtx, depID)
			if err != nil {
				fmt.Printf("Error checking dependency %s for task %s: %v\n", depID, task.ID, err)
				return
			}
			if len(results) == 0 {
				fmt.Printf("No results found for dependency %s of task %s\n", depID, task.ID)
				return
			}

			lastResult := results[len(results)-1]
			if lastResult.Status != models.TaskStatusCompleted {
				fmt.Printf("Dependency %s of task %s is not completed (status: %s)\n", depID, task.ID, lastResult.Status)
				return
			}
		}
		fmt.Printf("All dependencies completed for task %s\n", task.ID)
	}

	// Execute task
	fmt.Printf("Executing task %s\n", task.ID)
	result, err := s.worker.ExecuteTask(taskCtx, task)
	if err != nil {
		fmt.Printf("Error executing task %s: %v\n", task.ID, err)
		// Create a failed result if execution failed
		result = &models.TaskResult{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Status:    models.TaskStatusFailed,
			Error:     err.Error(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Version:   task.Version,
		}
		if err := s.storage.CreateTaskResult(taskCtx, result); err != nil {
			fmt.Printf("Error storing failed result for task %s: %v\n", task.ID, err)
		}
		return
	}
	fmt.Printf("Task %s executed successfully with status %s\n", task.ID, result.Status)

	// Update task status based on the result
	task.Status = result.Status
	task.UpdatedAt = time.Now()
	if err := s.storage.UpdateTask(taskCtx, task); err != nil {
		fmt.Printf("Error updating task %s status: %v\n", task.ID, err)
		return
	}
	fmt.Printf("Task %s status updated to %s\n", task.ID, task.Status)
}

// RemoveTask removes a task from the scheduler
func (s *Scheduler) RemoveTask(ctx context.Context, taskID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.entries[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.entries, taskID)
	}

	return nil
}

// ExecuteTaskNow executes a task immediately without scheduling
func (s *Scheduler) ExecuteTaskNow(ctx context.Context, task *models.Task) error {
	fmt.Printf("Starting immediate execution of task %s\n", task.ID)
	
	// Create a copy of the task to avoid race conditions
	taskCopy := *task
	
	// Create a new context with timeout
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Execute task directly and wait for result
	fmt.Printf("Executing task %s with worker\n", task.ID)
	result, err := s.worker.ExecuteTask(execCtx, &taskCopy)
	if err != nil {
		fmt.Printf("Error executing task %s: %v\n", task.ID, err)
		// Create a failed result if execution failed
		result = &models.TaskResult{
			ID:        uuid.New(),
			TaskID:    task.ID,
			Status:    models.TaskStatusFailed,
			Error:     err.Error(),
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Version:   task.Version,
		}
		fmt.Printf("Storing failed result for task %s\n", task.ID)
		if err := s.storage.CreateTaskResult(execCtx, result); err != nil {
			fmt.Printf("Error storing failed result for task %s: %v\n", task.ID, err)
			return fmt.Errorf("failed to store failed result: %w", err)
		}
		return fmt.Errorf("failed to execute task: %w", err)
	}

	// Store the successful result
	fmt.Printf("Task %s executed successfully, storing result\n", task.ID)
	if err := s.storage.CreateTaskResult(execCtx, result); err != nil {
		fmt.Printf("Error storing successful result for task %s: %v\n", task.ID, err)
		return fmt.Errorf("failed to store task result: %w", err)
	}
	fmt.Printf("Result stored successfully for task %s\n", task.ID)

	// Update task status based on the result
	task.Status = result.Status
	task.UpdatedAt = time.Now()
	fmt.Printf("Updating task %s status to %s\n", task.ID, task.Status)
	if err := s.storage.UpdateTask(execCtx, task); err != nil {
		fmt.Printf("Error updating task %s status: %v\n", task.ID, err)
		return fmt.Errorf("failed to update task status: %w", err)
	}
	fmt.Printf("Task %s status updated successfully\n", task.ID)

	return nil
} 