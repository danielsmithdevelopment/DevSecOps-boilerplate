package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/wallet-auth/internal/api"
	"github.com/yourusername/wallet-auth/internal/siwe"
	"github.com/yourusername/wallet-auth/internal/storage"
)

func main() {
	// Parse command line flags
	dbConnStr := flag.String("db-conn", os.Getenv("DATABASE_URL"), "Database connection string")
	jwtSecret := flag.String("jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret key")
	domain := flag.String("domain", os.Getenv("DOMAIN"), "Domain for SIWE messages")
	uri := flag.String("uri", os.Getenv("URI"), "URI for SIWE messages")
	chainID := flag.Int64("chain-id", 1, "Chain ID for SIWE messages")
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	// Validate required parameters
	if *dbConnStr == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if *jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if *domain == "" {
		log.Fatal("DOMAIN is required")
	}
	if *uri == "" {
		*uri = "https://" + *domain
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize storage
	store, err := storage.NewPostgresStorage(*dbConnStr)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize schema
	if err := store.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize SIWE verifier
	verifier := siwe.NewVerifier(*domain, *uri, *chainID)

	// Initialize API
	apiInstance := api.NewAPI(store, verifier, []byte(*jwtSecret))

	// Start API server
	go func() {
		log.Printf("Starting server on %s", *addr)
		if err := apiInstance.Run(*addr); err != nil {
			log.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
