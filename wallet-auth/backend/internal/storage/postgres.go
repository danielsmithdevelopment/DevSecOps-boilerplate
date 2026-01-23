package storage

import (
	"context"
	"database/sql"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/yourusername/wallet-auth/internal/models"
)

// Storage defines the database interface
type Storage interface {
	GetOrCreateUser(ctx context.Context, walletAddress string) (*models.User, error)
	UpdateUserNonce(ctx context.Context, userID uuid.UUID, nonce string) error
	GetUserByAddress(ctx context.Context, walletAddress string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	InitSchema(ctx context.Context) error
	Close() error
}

// PostgresStorage implements Storage using PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgreSQL storage instance
func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// InitSchema initializes the database schema
func (s *PostgresStorage) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		wallet_address VARCHAR(42) UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT NOW(),
		last_login TIMESTAMP,
		nonce VARCHAR(32) NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_users_wallet_address ON users(wallet_address);
	`

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// generateNonce generates a random nonce
func generateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetOrCreateUser gets an existing user or creates a new one
func (s *PostgresStorage) GetOrCreateUser(ctx context.Context, walletAddress string) (*models.User, error) {
	// Normalize address to lowercase
	walletAddress = strings.ToLower(walletAddress)
	
	// Try to get existing user
	user, err := s.GetUserByAddress(ctx, walletAddress)
	if err == nil {
		return user, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new user
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	query := `
		INSERT INTO users (wallet_address, nonce)
		VALUES ($1, $2)
		RETURNING id, wallet_address, created_at, last_login, nonce
	`

	user = &models.User{}
	err = s.db.QueryRowContext(ctx, query, walletAddress, nonce).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByAddress retrieves a user by wallet address
func (s *PostgresStorage) GetUserByAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	// Normalize address to lowercase for case-insensitive lookup
	walletAddress = strings.ToLower(strings.TrimSpace(walletAddress))
	log.Printf("DEBUG: GetUserByAddress called with normalized address: %q", walletAddress)
	query := `
		SELECT id, wallet_address, created_at, last_login, nonce
		FROM users
		WHERE LOWER(TRIM(wallet_address)) = LOWER(TRIM($1))
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		log.Printf("DEBUG: GetUserByAddress query failed for %q: %v", walletAddress, err)
		fmt.Fprintf(os.Stderr, "DEBUG: GetUserByAddress query failed for %q: %v\n", walletAddress, err)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("USER_NOT_FOUND_STORAGE: user not found for address %s", walletAddress)
		}
		return nil, fmt.Errorf("failed to get user by address %s: %w", walletAddress, err)
	}
	log.Printf("DEBUG: GetUserByAddress found user: %s (stored address: %q)", user.ID, user.WalletAddress)

	return user, nil
}

// UpdateUserNonce updates the nonce for a user
func (s *PostgresStorage) UpdateUserNonce(ctx context.Context, userID uuid.UUID, nonce string) error {
	query := `UPDATE users SET nonce = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, nonce, userID)
	return err
}

// UpdateLastLogin updates the last login timestamp for a user
func (s *PostgresStorage) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
