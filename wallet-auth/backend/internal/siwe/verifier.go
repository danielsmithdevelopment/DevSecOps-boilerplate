package siwe

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Verifier handles SIWE message generation and verification
type Verifier struct {
	domain   string
	uri      string
	chainID  int64
}

// NewVerifier creates a new SIWE verifier
func NewVerifier(domain, uri string, chainID int64) *Verifier {
	return &Verifier{
		domain:  domain,
		uri:     uri,
		chainID: chainID,
	}
}

// GenerateMessage generates a SIWE message for the given address and nonce
func (v *Verifier) GenerateMessage(address, nonce string) string {
	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	message := fmt.Sprintf(`%s wants you to sign in with your Ethereum account:
%s

URI: %s
Version: 1
Chain ID: %d
Nonce: %s
Issued At: %s`, v.domain, address, v.uri, v.chainID, nonce, timestamp)

	return message
}

// VerifySignature verifies an EIP-191 signature for a SIWE message
func (v *Verifier) VerifySignature(message, signature string) (string, error) {
	// Remove 0x prefix if present
	sig := strings.TrimPrefix(signature, "0x")
	if len(sig) != 130 {
		return "", fmt.Errorf("invalid signature length")
	}

	// Decode signature
	sigBytes := make([]byte, 65)
	_, err := hex.Decode(sigBytes, []byte(sig))
	if err != nil {
		return "", fmt.Errorf("failed to decode signature: %w", err)
	}

	// EIP-191: add prefix
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	prefixedMessage := append([]byte(prefix), []byte(message)...)
	hash := crypto.Keccak256Hash(prefixedMessage)

	// Recover public key
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}
	pubkey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get address from public key
	address := crypto.PubkeyToAddress(*pubkey)
	// Normalize to lowercase for consistent comparison
	return strings.ToLower(address.Hex()), nil
}

// ExtractAddressFromMessage extracts the Ethereum address from a SIWE message
func ExtractAddressFromMessage(message string) (string, error) {
	lines := strings.Split(message, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("invalid SIWE message format")
	}

	// Second line should contain the address
	addressLine := strings.TrimSpace(lines[1])
	if !common.IsHexAddress(addressLine) {
		return "", fmt.Errorf("invalid address in SIWE message")
	}

	// Normalize to lowercase for consistent storage/lookup
	address := common.HexToAddress(addressLine).Hex()
	return strings.ToLower(address), nil
}

// ExtractNonceFromMessage extracts the nonce from a SIWE message
func ExtractNonceFromMessage(message string) (string, error) {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Nonce: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Nonce: ")), nil
		}
	}
	return "", fmt.Errorf("nonce not found in SIWE message")
}
