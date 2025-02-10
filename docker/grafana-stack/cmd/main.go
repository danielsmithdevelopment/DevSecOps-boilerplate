package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	dataIngested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "data_ingested_total",
			Help: "Total number of data entries ingested",
		},
	)
)

func init() {
	prometheus.MustRegister(dataIngested)
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	// Get database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Use connection parameters if DATABASE_URL is not set
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("DB_HOST"),
			5432,
			os.Getenv("POSTGRES_DB"))
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database",
			zap.Error(err),
			zap.String("database_url", dbURL))
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		logger.Fatal("Failed to ping database",
			zap.Error(err))
	}
	logger.Info("Successfully connected to database")

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("Metrics server failed",
				zap.Error(err))
		}
	}()

	// Create table if not exists
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS my_table (
            id SERIAL PRIMARY KEY,
            data TEXT,
            timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		logger.Fatal("Failed to create table",
			zap.Error(err))
	}

	// Insert data
	_, err = db.Exec("INSERT INTO my_table (data) VALUES ($1)", "Hello, World!")
	if err != nil {
		logger.Fatal("Failed to insert data",
			zap.Error(err))
	}

	// Increment metrics counter
	dataIngested.Inc()

	logger.Info("Data inserted successfully",
		zap.Time("timestamp", time.Now()))

	// Keep the application running
	select {} // This replaces os.Exit(0) to keep the metrics endpoint available
}
