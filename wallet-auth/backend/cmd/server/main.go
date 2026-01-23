package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/wallet-auth/internal/api"
	"github.com/yourusername/wallet-auth/internal/email"
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
	
	// AWS SES configuration
	awsRegion := flag.String("aws-region", os.Getenv("AWS_REGION"), "AWS region for SES")
	awsAccessKeyID := flag.String("aws-access-key-id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key ID")
	awsSecretAccessKey := flag.String("aws-secret-access-key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key")
	sesFromEmail := flag.String("ses-from-email", os.Getenv("SES_FROM_EMAIL"), "SES sender email address")
	baseURL := flag.String("base-url", os.Getenv("BASE_URL"), "Base URL for email links")
	
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

	// Initialize email service (optional - can be nil if AWS credentials not provided)
	var emailService *email.EmailService
	if *awsRegion != "" && *sesFromEmail != "" {
		if *baseURL == "" {
			*baseURL = *uri
		}
		emailService, err = email.NewEmailService(*awsRegion, *awsAccessKeyID, *awsSecretAccessKey, *sesFromEmail, *baseURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize email service: %v. Email features will be disabled.", err)
			emailService = nil
		} else {
			log.Printf("Email service initialized successfully")
		}
	} else {
		log.Printf("AWS SES credentials not provided. Email features will be disabled.")
	}

	// Initialize API
	apiInstance := api.NewAPI(store, verifier, emailService, []byte(*jwtSecret))

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
