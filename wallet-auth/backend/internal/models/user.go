package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	WalletAddress          *string    `json:"wallet_address,omitempty" db:"wallet_address"`
	Email                  *string    `json:"email,omitempty" db:"email"`
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"`
	EmailVerificationExpires *time.Time `json:"-" db:"email_verification_expires"`
	WalletVerificationNonce *string    `json:"-" db:"wallet_verification_nonce"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	LastLogin              *time.Time `json:"last_login,omitempty" db:"last_login"`
	Nonce                  string     `json:"-" db:"nonce"`
}
