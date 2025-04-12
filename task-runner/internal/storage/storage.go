package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yourusername/task-runner/internal/models"
)

// Storage defines the interface for task storage
type Storage interface {
	// Task operations
	CreateTask(ctx context.Context, task *models.Task) error
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id uuid.UUID) error
	GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListTasks(ctx context.Context) ([]*models.Task, error)
	
	// Task result operations
	CreateTaskResult(ctx context.Context, result *models.TaskResult) error
	GetTaskResult(ctx context.Context, id uuid.UUID) (*models.TaskResult, error)
	ListTaskResults(ctx context.Context, taskID uuid.UUID) ([]*models.TaskResult, error)
	
	// Cleanup
	CleanupOldResults(ctx context.Context, olderThan time.Duration) error
}

// PostgresStorage implements the Storage interface using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

// DB returns the underlying database connection
func (s *PostgresStorage) DB() *sql.DB {
	return s.db
}

// InitSchema creates the necessary database tables
func (s *PostgresStorage) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			schedule VARCHAR(100) NOT NULL,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			created_by VARCHAR(255) NOT NULL,
			version INTEGER NOT NULL,
			config JSONB NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS task_dependencies (
			task_id UUID REFERENCES tasks(id),
			dependency_id UUID REFERENCES tasks(id),
			PRIMARY KEY (task_id, dependency_id)
		)`,
		`CREATE TABLE IF NOT EXISTS task_results (
			id UUID PRIMARY KEY,
			task_id UUID REFERENCES tasks(id),
			status VARCHAR(50) NOT NULL,
			output TEXT,
			error TEXT,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			version INTEGER NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

// CreateTask implements the Storage interface
func (s *PostgresStorage) CreateTask(ctx context.Context, task *models.Task) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Convert task config to JSON
	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal task config: %w", err)
	}

	query := `
		INSERT INTO tasks (id, name, type, schedule, status, created_at, updated_at, created_by, version, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.ExecContext(ctx, query,
		task.ID,
		task.Name,
		task.Type,
		task.Schedule,
		task.Status,
		task.CreatedAt,
		task.UpdatedAt,
		task.CreatedBy,
		task.Version,
		configJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	// Insert dependencies if any
	if len(task.Dependencies) > 0 {
		depQuery := `
			INSERT INTO task_dependencies (task_id, dependency_id)
			VALUES ($1, $2)
		`
		for _, depID := range task.Dependencies {
			if _, err := tx.ExecContext(ctx, depQuery, task.ID, depID); err != nil {
				return fmt.Errorf("failed to insert task dependency: %w", err)
			}
		}
	}

	return tx.Commit()
}

// CreateTaskResult implements the Storage interface
func (s *PostgresStorage) CreateTaskResult(ctx context.Context, result *models.TaskResult) error {
	query := `
		INSERT INTO task_results (id, task_id, status, output, error, start_time, end_time, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.ExecContext(ctx, query,
		result.ID,
		result.TaskID,
		result.Status,
		result.Output,
		result.Error,
		result.StartTime,
		result.EndTime,
		result.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to create task result: %w", err)
	}
	return nil
}

// CleanupOldResults implements the Storage interface
func (s *PostgresStorage) CleanupOldResults(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM task_results
		WHERE end_time < $1
	`
	cutoffTime := time.Now().Add(-olderThan)
	_, err := s.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old results: %w", err)
	}
	return nil
}

// GetTaskResult implements the Storage interface
func (s *PostgresStorage) GetTaskResult(ctx context.Context, id uuid.UUID) (*models.TaskResult, error) {
	query := `
		SELECT id, task_id, status, output, error, start_time, end_time, version
		FROM task_results
		WHERE id = $1
	`
	result := &models.TaskResult{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&result.ID,
		&result.TaskID,
		&result.Status,
		&result.Output,
		&result.Error,
		&result.StartTime,
		&result.EndTime,
		&result.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task result not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task result: %w", err)
	}
	return result, nil
}

// ListTaskResults implements the Storage interface
func (s *PostgresStorage) ListTaskResults(ctx context.Context, taskID uuid.UUID) ([]*models.TaskResult, error) {
	query := `
		SELECT id, task_id, status, output, error, start_time, end_time, version
		FROM task_results
		WHERE task_id = $1
		ORDER BY start_time DESC
	`
	rows, err := s.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list task results: %w", err)
	}
	defer rows.Close()

	var results []*models.TaskResult
	for rows.Next() {
		result := &models.TaskResult{}
		err := rows.Scan(
			&result.ID,
			&result.TaskID,
			&result.Status,
			&result.Output,
			&result.Error,
			&result.StartTime,
			&result.EndTime,
			&result.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task result: %w", err)
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task results: %w", err)
	}
	return results, nil
}

// GetTask implements the Storage interface
func (s *PostgresStorage) GetTask(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	task := &models.Task{}
	var configJSON []byte

	query := `
		SELECT id, name, type, schedule, status, created_at, updated_at, created_by, version, config
		FROM tasks
		WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Name,
		&task.Type,
		&task.Schedule,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CreatedBy,
		&task.Version,
		&configJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Deserialize task config from JSON
	if err := json.Unmarshal(configJSON, &task.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task config: %w", err)
	}

	// Get dependencies
	depQuery := `
		SELECT dependency_id
		FROM task_dependencies
		WHERE task_id = $1
	`
	rows, err := s.db.QueryContext(ctx, depQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []uuid.UUID
	for rows.Next() {
		var depID uuid.UUID
		if err := rows.Scan(&depID); err != nil {
			return nil, fmt.Errorf("failed to scan dependency ID: %w", err)
		}
		dependencies = append(dependencies, depID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dependencies: %w", err)
	}
	task.Dependencies = dependencies

	return task, nil
}

// UpdateTask implements the Storage interface
func (s *PostgresStorage) UpdateTask(ctx context.Context, task *models.Task) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize task config to JSON
	configJSON, err := json.Marshal(task.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal task config: %w", err)
	}

	query := `
		UPDATE tasks
		SET name = $1, type = $2, schedule = $3, status = $4, updated_at = $5, version = $6, config = $7
		WHERE id = $8
	`
	result, err := tx.ExecContext(ctx, query,
		task.Name,
		task.Type,
		task.Schedule,
		task.Status,
		time.Now(),
		task.Version,
		configJSON,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Delete existing dependencies
	_, err = tx.ExecContext(ctx, "DELETE FROM task_dependencies WHERE task_id = $1", task.ID)
	if err != nil {
		return fmt.Errorf("failed to delete existing dependencies: %w", err)
	}

	// Insert new dependencies
	if len(task.Dependencies) > 0 {
		depQuery := `
			INSERT INTO task_dependencies (task_id, dependency_id)
			VALUES ($1, $2)
		`
		for _, depID := range task.Dependencies {
			_, err = tx.ExecContext(ctx, depQuery, task.ID, depID)
			if err != nil {
				return fmt.Errorf("failed to insert dependency: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteTask implements the Storage interface
func (s *PostgresStorage) DeleteTask(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete dependencies first
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_dependencies WHERE task_id = $1", id); err != nil {
		return fmt.Errorf("failed to delete dependencies: %w", err)
	}

	// Delete task results
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_results WHERE task_id = $1", id); err != nil {
		return fmt.Errorf("failed to delete task results: %w", err)
	}

	// Delete the task
	result, err := tx.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %s", id)
	}

	return tx.Commit()
}

// ListTasks implements the Storage interface
func (s *PostgresStorage) ListTasks(ctx context.Context) ([]*models.Task, error) {
	query := `
		SELECT id, name, type, schedule, status, created_at, updated_at, version, config
		FROM tasks
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var configJSON []byte
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Type,
			&task.Schedule,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.Version,
			&configJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Deserialize task config from JSON
		if err := json.Unmarshal(configJSON, &task.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task config: %w", err)
		}

		// Get dependencies for this task
		depQuery := `
			SELECT dependency_id
			FROM task_dependencies
			WHERE task_id = $1
		`
		depRows, err := s.db.QueryContext(ctx, depQuery, task.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query dependencies: %w", err)
		}
		defer depRows.Close()

		var dependencies []uuid.UUID
		for depRows.Next() {
			var depID uuid.UUID
			if err := depRows.Scan(&depID); err != nil {
				return nil, fmt.Errorf("failed to scan dependency: %w", err)
			}
			dependencies = append(dependencies, depID)
		}
		if err := depRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating dependencies: %w", err)
		}

		task.Dependencies = dependencies
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// Additional methods would be implemented here... 