package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"task-runner/storage"
	"task-runner/worker"
	"task-runner/scheduler"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get instance ID from environment variable
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "default"
	}
	log.Printf("Starting task-runner instance: %s", instanceID)

	// Initialize storage
	storage, err := storage.NewPostgresStorage(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Initialize worker
	worker := worker.NewWorker(storage)

	// Initialize scheduler
	scheduler := scheduler.NewScheduler(storage, worker, instanceID)

	// ... existing code ...
} 