package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/task-runner/internal/models"
	"github.com/yourusername/task-runner/internal/scheduler"
	"github.com/yourusername/task-runner/internal/storage"
)

// API handles HTTP endpoints
type API struct {
	storage   storage.Storage
	scheduler *scheduler.Scheduler
	router    *gin.Engine
	jwtSecret []byte
	startTime time.Time
}

// NewAPI creates a new API instance
func NewAPI(storage storage.Storage, scheduler *scheduler.Scheduler, jwtSecret []byte) *API {
	api := &API{
		storage:   storage,
		scheduler: scheduler,
		router:    gin.Default(),
		jwtSecret: jwtSecret,
		startTime: time.Now(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures the API routes
func (a *API) setupRoutes() {
	// Public routes
	a.router.POST("/auth", a.handleAuth)
	a.router.GET("/health", a.handleHealth)
	a.router.GET("/metrics", a.handleMetrics)

	// Protected routes
	protected := a.router.Group("/")
	protected.Use(a.authMiddleware())
	{
		protected.POST("/task/create", a.handleCreateTask)
		protected.PUT("/task/update", a.handleUpdateTask)
		protected.DELETE("/task/delete/:id", a.handleDeleteTask)
		protected.POST("/task/invoke/:id", a.handleInvokeTask)
		protected.GET("/task/status/:id", a.handleTaskStatus)
		protected.GET("/task/list", a.handleListTasks)
	}
}

// handleMetrics returns Prometheus-formatted metrics
func (a *API) handleMetrics(c *gin.Context) {
	// Get task results for the last hour
	ctx := c.Request.Context()
	results, err := a.storage.ListTaskResults(ctx, uuid.Nil) // This will get all results
	if err != nil {
		fmt.Printf("Error getting task results: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task results"})
		return
	}

	fmt.Printf("Found %d task results\n", len(results))

	// Count tasks by instance and status
	instanceTaskCounts := make(map[string]map[string]int)
	instanceDurations := make(map[string][]float64)

	for _, result := range results {
		instanceID := result.Metadata["instance_id"]
		if instanceID == "" {
			instanceID = "unknown"
		}

		fmt.Printf("Processing result for instance %s with status %s\n", instanceID, result.Status)

		// Initialize maps if needed
		if _, ok := instanceTaskCounts[instanceID]; !ok {
			instanceTaskCounts[instanceID] = make(map[string]int)
		}

		// Count tasks by status
		instanceTaskCounts[instanceID][string(result.Status)]++

		// Calculate duration
		duration := result.EndTime.Sub(result.StartTime).Seconds()
		instanceDurations[instanceID] = append(instanceDurations[instanceID], duration)
	}

	// Build metrics output
	metrics := fmt.Sprintf(`# HELP task_runner_info Information about the task runner service
# TYPE task_runner_info gauge
task_runner_info{version="1.0.0"} 1

# HELP task_runner_uptime_seconds The number of seconds since the service started
# TYPE task_runner_uptime_seconds gauge
task_runner_uptime_seconds %d

# HELP task_runner_tasks_total Total number of tasks executed by instance and status
# TYPE task_runner_tasks_total counter
`, int(time.Since(a.startTime).Seconds()))

	// Add task counts by instance and status
	for instanceID, statusCounts := range instanceTaskCounts {
		for status, count := range statusCounts {
			metrics += fmt.Sprintf("task_runner_tasks_total{instance=\"%s\",status=\"%s\"} %d\n", 
				instanceID, status, count)
		}
	}

	// Add average duration by instance
	metrics += "\n# HELP task_runner_task_duration_seconds Average task execution duration by instance\n"
	metrics += "# TYPE task_runner_task_duration_seconds gauge\n"
	for instanceID, durations := range instanceDurations {
		var total float64
		for _, d := range durations {
			total += d
		}
		avg := total / float64(len(durations))
		metrics += fmt.Sprintf("task_runner_task_duration_seconds{instance=\"%s\"} %.2f\n", 
			instanceID, avg)
	}

	fmt.Printf("Generated metrics:\n%s\n", metrics)

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, metrics)
}

// authMiddleware handles JWT authentication
func (a *API) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		// Format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		token := parts[1]

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return a.jwtSecret, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}

// handleAuth handles user authentication
func (a *API) handleAuth(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// TODO: Implement actual authentication
	// For now, just generate a token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  req.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// handleCreateTask handles task creation
func (a *API) handleCreateTask(c *gin.Context) {
	var task models.Task
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task format"})
		return
	}

	// Generate UUID for new task
	task.ID = uuid.New()
	userID := c.GetString("user_id")
	task.CreatedBy = userID
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = models.TaskStatusPending
	task.Version = 1

	if err := a.storage.CreateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	if err := a.scheduler.ScheduleTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleUpdateTask handles task updates
func (a *API) handleUpdateTask(c *gin.Context) {
	var task models.Task
	if err := c.BindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task format"})
		return
	}

	task.UpdatedAt = time.Now()
	task.Version++

	if err := a.storage.UpdateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update task"})
		return
	}

	if err := a.scheduler.ScheduleTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reschedule task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// handleDeleteTask handles task deletion
func (a *API) handleDeleteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	if err := a.scheduler.RemoveTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove task from scheduler"})
		return
	}

	if err := a.storage.DeleteTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete task"})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleInvokeTask handles manual task invocation
func (a *API) handleInvokeTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := a.storage.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	if err := a.scheduler.ExecuteTaskNow(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to invoke task"})
		return
	}

	c.Status(http.StatusAccepted)
}

// handleTaskStatus handles task status retrieval
func (a *API) handleTaskStatus(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	task, err := a.storage.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	results, err := a.storage.ListTaskResults(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task":    task,
		"results": results,
	})
}

// handleHealth handles health check requests
func (a *API) handleHealth(c *gin.Context) {
	// Check database connection
	if err := a.storage.(*storage.PostgresStorage).DB().Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// handleListTasks handles listing all tasks
func (a *API) handleListTasks(c *gin.Context) {
	tasks, err := a.storage.ListTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// Run starts the API server
func (a *API) Run(addr string) error {
	return a.router.Run(addr)
} 