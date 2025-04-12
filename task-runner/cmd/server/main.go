package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/task-runner/internal/api"
	"github.com/yourusername/task-runner/internal/scheduler"
	"github.com/yourusername/task-runner/internal/storage"
	"github.com/yourusername/task-runner/internal/worker"
)

func main() {
	// Parse command line flags
	dbConnStr := flag.String("db-conn", os.Getenv("DB_CONN"), "Database connection string")
	jwtSecret := flag.String("jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret key")
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	store, err := storage.NewPostgresStorage(*dbConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize schema
	if err := store.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize worker
	worker := worker.NewWorker(store)

	// Initialize scheduler
	scheduler := scheduler.NewScheduler(store, worker)

	// Start scheduler
	if err := scheduler.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	defer scheduler.Stop()

	// Initialize API
	api := api.NewAPI(store, scheduler, []byte(*jwtSecret))

	// Start API server
	go func() {
		if err := api.Run(*addr); err != nil {
			log.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("Shutting down...")
	cancel()

	// Give some time for cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// TODO: Implement graceful shutdown for API server
	<-shutdownCtx.Done()
	log.Println("Shutdown complete")
} 