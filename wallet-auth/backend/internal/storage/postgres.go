package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

	// Email-related methods
	GetOrCreateUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateEmailVerification(ctx context.Context, userID uuid.UUID, token string, expires time.Time) error
	VerifyEmail(ctx context.Context, token string) (*models.User, error)
	UpdateWalletVerificationNonce(ctx context.Context, userID uuid.UUID, nonce string) error
	LinkWalletToUser(ctx context.Context, userID uuid.UUID, address string) error
	LinkEmailToUser(ctx context.Context, userID uuid.UUID, email string) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	SetEmailUnverified(ctx context.Context, userID uuid.UUID, email string) error
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
	// Create table if it doesn't exist
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		wallet_address VARCHAR(42) UNIQUE,
		created_at TIMESTAMP DEFAULT NOW(),
		last_login TIMESTAMP,
		nonce VARCHAR(32) NOT NULL
	);
	`
	if _, err := s.db.ExecContext(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Migrate existing schema: make wallet_address nullable
	// This will fail if column is already nullable, which is fine
	migrateWalletQuery := `
	DO $$ 
	BEGIN
		ALTER TABLE users ALTER COLUMN wallet_address DROP NOT NULL;
	EXCEPTION
		WHEN others THEN NULL;
	END $$;
	`
	if _, err := s.db.ExecContext(ctx, migrateWalletQuery); err != nil {
		log.Printf("Warning: Could not make wallet_address nullable (may already be nullable): %v", err)
	}

	// Add email columns if they don't exist
	migrateEmailQuery := `
	DO $$ 
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email') THEN
			ALTER TABLE users ADD COLUMN email VARCHAR(255) UNIQUE;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email_verified') THEN
			ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email_verification_token') THEN
			ALTER TABLE users ADD COLUMN email_verification_token VARCHAR(64);
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email_verification_expires') THEN
			ALTER TABLE users ADD COLUMN email_verification_expires TIMESTAMP;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='wallet_verification_nonce') THEN
			ALTER TABLE users ADD COLUMN wallet_verification_nonce VARCHAR(32);
		END IF;
	END $$;
	`
	if _, err := s.db.ExecContext(ctx, migrateEmailQuery); err != nil {
		return fmt.Errorf("failed to add email columns: %w", err)
	}

	// Create indexes
	indexQueries := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_wallet_address ON users(wallet_address) WHERE wallet_address IS NOT NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users(email_verification_token) WHERE email_verification_token IS NOT NULL;`,
	}

	for _, query := range indexQueries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			log.Printf("Warning: Could not create index: %v", err)
		}
	}

	return nil
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
		RETURNING id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
	`

	user = &models.User{}
	err = s.db.QueryRowContext(ctx, query, walletAddress, nonce).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
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
		SELECT id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
		FROM users
		WHERE LOWER(TRIM(wallet_address)) = LOWER(TRIM($1))
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
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
	log.Printf("DEBUG: GetUserByAddress found user: %s (stored address: %v)", user.ID, user.WalletAddress)

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

// GetUserByID retrieves a user by ID
func (s *PostgresStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for id %s", userID)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

// GetOrCreateUserByEmail gets an existing user or creates a new one by email
func (s *PostgresStorage) GetOrCreateUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Normalize email to lowercase
	email = strings.ToLower(strings.TrimSpace(email))

	// Try to get existing user
	user, err := s.GetUserByEmail(ctx, email)
	if err == nil {
		return user, nil
	}
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, err
	}

	// Create new user
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	query := `
		INSERT INTO users (email, nonce, email_verified)
		VALUES ($1, $2, FALSE)
		RETURNING id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
	`

	user = &models.User{}
	err = s.db.QueryRowContext(ctx, query, email, nonce).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Normalize email to lowercase
	email = strings.ToLower(strings.TrimSpace(email))

	query := `
		SELECT id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
		FROM users
		WHERE LOWER(TRIM(email)) = LOWER(TRIM($1))
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for email %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateEmailVerification stores the verification token and expiration
func (s *PostgresStorage) UpdateEmailVerification(ctx context.Context, userID uuid.UUID, token string, expires time.Time) error {
	query := `
		UPDATE users 
		SET email_verification_token = $1, email_verification_expires = $2
		WHERE id = $3
	`
	_, err := s.db.ExecContext(ctx, query, token, expires, userID)
	return err
}

// VerifyEmail verifies the email token and marks email as verified
func (s *PostgresStorage) VerifyEmail(ctx context.Context, token string) (*models.User, error) {
	query := `
		SELECT id, wallet_address, email, email_verified, email_verification_token, email_verification_expires, wallet_verification_nonce, created_at, last_login, nonce
		FROM users
		WHERE email_verification_token = $1 AND email_verification_expires > NOW()
	`

	user := &models.User{}
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Email,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.EmailVerificationExpires,
		&user.WalletVerificationNonce,
		&user.CreatedAt,
		&user.LastLogin,
		&user.Nonce,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired verification token")
		}
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Mark email as verified and clear token
	updateQuery := `
		UPDATE users 
		SET email_verified = TRUE, email_verification_token = NULL, email_verification_expires = NULL
		WHERE id = $1
	`
	if _, err := s.db.ExecContext(ctx, updateQuery, user.ID); err != nil {
		return nil, fmt.Errorf("failed to update email verification status: %w", err)
	}

	user.EmailVerified = true
	user.EmailVerificationToken = nil
	user.EmailVerificationExpires = nil

	return user, nil
}

// UpdateWalletVerificationNonce stores a nonce for wallet verification when adding wallet to email user
func (s *PostgresStorage) UpdateWalletVerificationNonce(ctx context.Context, userID uuid.UUID, nonce string) error {
	query := `UPDATE users SET wallet_verification_nonce = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, nonce, userID)
	return err
}

// LinkWalletToUser links a wallet address to an existing user (for email-first users)
func (s *PostgresStorage) LinkWalletToUser(ctx context.Context, userID uuid.UUID, address string) error {
	// Normalize address
	address = strings.ToLower(strings.TrimSpace(address))

	// Check if address is already linked to another user
	existingUser, err := s.GetUserByAddress(ctx, address)
	if err == nil && existingUser.ID != userID {
		return fmt.Errorf("wallet address already linked to another user")
	}

	query := `
		UPDATE users 
		SET wallet_address = $1, wallet_verification_nonce = NULL
		WHERE id = $2
	`
	_, err = s.db.ExecContext(ctx, query, address, userID)
	return err
}

// SetEmailUnverified sets the email as unverified (for adding email to wallet user)
func (s *PostgresStorage) SetEmailUnverified(ctx context.Context, userID uuid.UUID, emailAddr string) error {
	// Normalize email
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))

	query := `
		UPDATE users 
		SET email = $1, email_verified = FALSE
		WHERE id = $2
	`
	_, err := s.db.ExecContext(ctx, query, emailAddr, userID)
	return err
}

// LinkEmailToUser links an email to an existing user (for wallet-first users)
func (s *PostgresStorage) LinkEmailToUser(ctx context.Context, userID uuid.UUID, email string) error {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Check if email is already linked to another user
	existingUser, err := s.GetUserByEmail(ctx, email)
	if err == nil && existingUser.ID != userID {
		return fmt.Errorf("email already linked to another user")
	}

	query := `
		UPDATE users 
		SET email = $1, email_verified = TRUE, email_verification_token = NULL, email_verification_expires = NULL
		WHERE id = $2
	`
	_, err = s.db.ExecContext(ctx, query, email, userID)
	return err
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
