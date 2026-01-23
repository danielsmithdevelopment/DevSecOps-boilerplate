package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/wallet-auth/internal/siwe"
	"github.com/yourusername/wallet-auth/internal/storage"
)

// API handles HTTP endpoints
type API struct {
	storage  storage.Storage
	verifier *siwe.Verifier
	jwtSecret []byte
	router   *gin.Engine
}

// NewAPI creates a new API instance
func NewAPI(storage storage.Storage, verifier *siwe.Verifier, jwtSecret []byte) *API {
	api := &API{
		storage:  storage,
		verifier: verifier,
		jwtSecret: jwtSecret,
		router:   gin.Default(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures API routes
func (a *API) setupRoutes() {
	// CORS middleware
	a.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	a.router.GET("/health", a.handleHealth)
	a.router.GET("/auth/challenge", a.handleChallenge)
	a.router.POST("/auth/verify", a.handleVerify)

	// Protected routes
	protected := a.router.Group("/")
	protected.Use(a.authMiddleware())
	{
		protected.GET("/auth/me", a.handleMe)
	}
}

// handleHealth returns health status
func (a *API) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleChallenge generates a SIWE challenge message
func (a *API) handleChallenge(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}
	// Normalize address to lowercase
	address = strings.ToLower(address)

	// Get or create user
	user, err := a.storage.GetOrCreateUser(c.Request.Context(), address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Generate new nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Update user nonce
	if err := a.storage.UpdateUserNonce(c.Request.Context(), user.ID, nonce); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update nonce"})
		return
	}

	// Generate SIWE message
	message := a.verifier.GenerateMessage(address, nonce)

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"nonce":   nonce,
	})
}

// handleVerify verifies the signature and returns a JWT token
func (a *API) handleVerify(c *gin.Context) {
	fmt.Fprintf(os.Stderr, "=== handleVerify CALLED ===\n")
	log.Printf("DEBUG: handleVerify called")
	var req struct {
		Message   string `json:"message" binding:"required"`
		Signature string `json:"signature" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		log.Printf("DEBUG: BindJSON failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	log.Printf("DEBUG: Received verify request, message length: %d", len(req.Message))

	// Extract address from message
	address, err := siwe.ExtractAddressFromMessage(req.Message)
	if err != nil {
		log.Printf("DEBUG: Failed to extract address: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid message format: %v", err)})
		return
	}
	// Address is already normalized to lowercase by ExtractAddressFromMessage
	// But ensure it's lowercase for consistency
	address = strings.ToLower(address)
	log.Printf("DEBUG: Extracted address from message: %q", address)

	// Extract nonce from message
	nonce, err := siwe.ExtractNonceFromMessage(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message format"})
		return
	}

	// Verify signature
	recoveredAddress, err := a.verifier.VerifySignature(req.Message, req.Signature)
	if err != nil {
		log.Printf("DEBUG: Signature verification failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid signature: %v", err)})
		return
	}
	log.Printf("DEBUG: Recovered address from signature: %q", recoveredAddress)

	// Check if recovered address matches
	if strings.ToLower(recoveredAddress) != strings.ToLower(address) {
		errorMsg := fmt.Sprintf("address mismatch: recovered %s (len:%d), expected %s (len:%d)", recoveredAddress, len(recoveredAddress), address, len(address))
		log.Printf("DEBUG: %s", errorMsg)
		c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
		return
	}

	// Get user using recovered address (authoritative)
	// The user should have been created during the challenge request
	lookupAddress := strings.ToLower(strings.TrimSpace(recoveredAddress))
	
	// Log for debugging
	log.Printf("DEBUG: Looking up user with address: %q (recovered: %q, from message: %q)", lookupAddress, recoveredAddress, address)
	
	user, err := a.storage.GetUserByAddress(c.Request.Context(), lookupAddress)
	if err != nil {
		// Log the error immediately
		fmt.Fprintf(os.Stderr, "ERROR: Primary lookup failed for %q: %v\n", lookupAddress, err)
		// If user not found, it might be because they used a different address format
		// Try to get user by the address from message as fallback
		fallbackAddress := strings.ToLower(strings.TrimSpace(address))
		fmt.Fprintf(os.Stderr, "ERROR: Attempting fallback lookup for %q\n", fallbackAddress)
		fallbackUser, fallbackErr := a.storage.GetUserByAddress(c.Request.Context(), fallbackAddress)
		if fallbackErr == nil {
			user = fallbackUser
			fmt.Fprintf(os.Stderr, "ERROR: Found user via fallback\n")
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: Fallback lookup also failed: %v\n", fallbackErr)
			// Make error message very explicit and include all details
			errorDetails := fmt.Sprintf("USER_NOT_FOUND: lookup_addr=%q(len=%d) recovered_addr=%q(len=%d) message_addr=%q(len=%d) primary_error=%v fallback_error=%v", 
				lookupAddress, len(lookupAddress), recoveredAddress, len(recoveredAddress), address, len(address), err, fallbackErr)
			fmt.Fprintf(os.Stderr, "ERROR: Returning error: %s\n", errorDetails)
			// Return detailed error with additional debug info
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errorDetails,
				"debug": gin.H{
					"lookup_address": lookupAddress,
					"recovered_address": recoveredAddress,
					"message_address": address,
					"primary_error": err.Error(),
					"fallback_error": fallbackErr.Error(),
				},
			})
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "SUCCESS: Found user with address %q\n", lookupAddress)
	}

	if user.Nonce != nonce {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("invalid nonce: got %s, expected %s", nonce, user.Nonce)})
		return
	}

	// Update last login
	if err := a.storage.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		// Log error but don't fail the request
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":       user.ID.String(),
		"wallet_address": user.WalletAddress,
		"exp":           time.Now().Add(time.Hour * 24).Unix(),
		"iat":           time.Now().Unix(),
	})

	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":             user.ID,
			"wallet_address": user.WalletAddress,
		},
	})
}

// handleMe returns the current authenticated user
func (a *API) handleMe(c *gin.Context) {
	walletAddress := c.GetString("wallet_address")
	fmt.Fprintf(os.Stderr, "=== handleMe: wallet_address from context = %q ===\n", walletAddress)

	user, err := a.storage.GetUserByAddress(c.Request.Context(), walletAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "=== handleMe: user lookup failed for %q: %v ===\n", walletAddress, err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("user not found: address=%q, err=%v", walletAddress, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"wallet_address": user.WalletAddress,
		"created_at":     user.CreatedAt,
		"last_login":     user.LastLogin,
	})
}

// Run starts the HTTP server
func (a *API) Run(addr string) error {
	return a.router.Run(addr)
}
