package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	WalletAddress string  `json:"wallet_address" db:"wallet_address"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
	Nonce        string    `json:"-" db:"nonce"`
}
