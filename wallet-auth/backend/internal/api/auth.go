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
	"github.com/google/uuid"
	"github.com/yourusername/wallet-auth/internal/email"
	"github.com/yourusername/wallet-auth/internal/siwe"
	"github.com/yourusername/wallet-auth/internal/storage"
)

// API handles HTTP endpoints
type API struct {
	storage      storage.Storage
	verifier     *siwe.Verifier
	emailService *email.EmailService
	jwtSecret    []byte
	router       *gin.Engine
}

// NewAPI creates a new API instance
func NewAPI(storage storage.Storage, verifier *siwe.Verifier, emailService *email.EmailService, jwtSecret []byte) *API {
	api := &API{
		storage:      storage,
		verifier:     verifier,
		emailService: emailService,
		jwtSecret:    jwtSecret,
		router:       gin.Default(),
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

	// Email auth routes
	a.router.POST("/auth/email/signup", a.handleEmailSignup)
	a.router.GET("/auth/email/verify", a.handleEmailVerify)
	a.router.POST("/auth/email/resend", a.handleEmailResend)

	// Protected routes
	protected := a.router.Group("/")
	protected.Use(a.authMiddleware())
	{
		protected.GET("/auth/me", a.handleMe)
		protected.POST("/auth/wallet/add", a.handleWalletAdd)
		protected.POST("/auth/wallet/verify", a.handleWalletVerify)
		protected.POST("/auth/email/add", a.handleEmailAdd)
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
	if !strings.EqualFold(recoveredAddress, address) {
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
					"lookup_address":    lookupAddress,
					"recovered_address": recoveredAddress,
					"message_address":   address,
					"primary_error":     err.Error(),
					"fallback_error":    fallbackErr.Error(),
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
		log.Printf("failed to update last login: %v", err)
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"auth_method": "wallet",
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
		"iat":         time.Now().Unix(),
	}
	if user.WalletAddress != nil {
		claims["wallet_address"] = *user.WalletAddress
	}
	if user.Email != nil {
		claims["email"] = *user.Email
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	userResponse := gin.H{
		"id": user.ID,
	}
	if user.WalletAddress != nil {
		userResponse["wallet_address"] = *user.WalletAddress
	}
	if user.Email != nil {
		userResponse["email"] = *user.Email
		userResponse["email_verified"] = user.EmailVerified
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  userResponse,
	})
}

// handleMe returns the current authenticated user
func (a *API) handleMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	user, err := a.storage.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	response := gin.H{
		"id":         user.ID,
		"created_at": user.CreatedAt,
		"last_login": user.LastLogin,
	}
	if user.WalletAddress != nil {
		response["wallet_address"] = *user.WalletAddress
	}
	if user.Email != nil {
		response["email"] = *user.Email
		response["email_verified"] = user.EmailVerified
	}

	c.JSON(http.StatusOK, response)
}

// handleEmailSignup handles email signup and sends verification email
func (a *API) handleEmailSignup(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	// Get or create user
	user, err := a.storage.GetOrCreateUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate verification token
	token, err := email.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	expires := email.GetVerificationExpiry()

	// Store verification token
	if err := a.storage.UpdateEmailVerification(c.Request.Context(), user.ID, token, expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store verification token"})
		return
	}

	// Send verification email
	emailSent := false
	if a.emailService != nil {
		if err := a.emailService.SendVerificationEmail(req.Email, token); err != nil {
			log.Printf("Failed to send verification email: %v", err)
			// Don't fail the request, just log the error
		} else {
			emailSent = true
		}
	}

	response := gin.H{
		"message":               "Verification email sent. Please check your inbox.",
		"email_sent":            emailSent,
		"email_service_enabled": a.emailService != nil,
	}

	// If email service is disabled, include the token so user can verify manually
	if a.emailService == nil {
		verificationURL := fmt.Sprintf("%s/email-verify?token=%s", os.Getenv("BASE_URL"), token)
		if verificationURL == "/email-verify?token="+token {
			// Fallback if BASE_URL not set
			verificationURL = fmt.Sprintf("http://localhost:3000/email-verify?token=%s", token)
		}
		response["verification_token"] = token
		response["verification_url"] = verificationURL
		response["message"] = "Email service is not configured. Use the verification link below to verify your email."
		log.Printf("Email service disabled - Verification token for %s: %s", req.Email, token)
	}

	c.JSON(http.StatusOK, response)
}

// handleEmailVerify verifies the email token and returns a JWT
func (a *API) handleEmailVerify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token parameter is required"})
		return
	}

	// Verify token
	user, err := a.storage.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// Update last login
	if err := a.storage.UpdateLastLogin(c.Request.Context(), user.ID); err != nil {
		log.Printf("failed to update last login: %v", err)
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"auth_method": "email",
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
		"iat":         time.Now().Unix(),
	}
	if user.Email != nil {
		claims["email"] = *user.Email
	}
	if user.WalletAddress != nil {
		claims["wallet_address"] = *user.WalletAddress
	}

	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := tokenJWT.SignedString(a.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":             user.ID,
			"email":          user.Email,
			"email_verified": user.EmailVerified,
		},
	})
}

// handleEmailResend resends the verification email
func (a *API) handleEmailResend(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	// Get user by email
	user, err := a.storage.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a verification link has been sent."})
		return
	}

	// Generate new verification token
	token, err := email.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	expires := email.GetVerificationExpiry()

	// Store verification token
	if err := a.storage.UpdateEmailVerification(c.Request.Context(), user.ID, token, expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store verification token"})
		return
	}

	// Send verification email
	emailSent := false
	if a.emailService != nil {
		if err := a.emailService.SendVerificationEmail(req.Email, token); err != nil {
			log.Printf("Failed to send verification email: %v", err)
		} else {
			emailSent = true
		}
	}

	response := gin.H{
		"message":               "If the email exists, a verification link has been sent.",
		"email_sent":            emailSent,
		"email_service_enabled": a.emailService != nil,
	}

	// If email service is disabled, include the token so user can verify manually
	if a.emailService == nil && user != nil {
		verificationURL := fmt.Sprintf("%s/email-verify?token=%s", os.Getenv("BASE_URL"), token)
		if verificationURL == "/email-verify?token="+token {
			// Fallback if BASE_URL not set
			verificationURL = fmt.Sprintf("http://localhost:3000/email-verify?token=%s", token)
		}
		response["verification_token"] = token
		response["verification_url"] = verificationURL
		response["message"] = "Email service is not configured. Use the verification link below if the email exists."
		log.Printf("Email service disabled - Resend verification token for %s: %s", req.Email, token)
	}

	c.JSON(http.StatusOK, response)
}

// handleWalletAdd initiates adding a wallet to an email user
func (a *API) handleWalletAdd(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Get user
	user, err := a.storage.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check if user has email verified
	if user.Email == nil || !user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email must be verified before adding wallet"})
		return
	}

	// Check if wallet already exists
	if user.WalletAddress != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet already linked"})
		return
	}

	var req struct {
		Address string `json:"address" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}

	// Normalize address
	address := strings.ToLower(strings.TrimSpace(req.Address))

	// Check if address is already linked to another user
	existingUser, err := a.storage.GetUserByAddress(c.Request.Context(), address)
	if err == nil && existingUser.ID != userID {
		c.JSON(http.StatusConflict, gin.H{"error": "wallet address already linked to another account"})
		return
	}

	// Generate nonce for wallet verification
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}
	nonce := hex.EncodeToString(nonceBytes)

	// Store verification nonce
	if err := a.storage.UpdateWalletVerificationNonce(c.Request.Context(), userID, nonce); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store verification nonce"})
		return
	}

	// Generate SIWE message
	message := a.verifier.GenerateMessage(address, nonce)

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"nonce":   nonce,
	})
}

// handleWalletVerify verifies wallet signature when adding to email user
func (a *API) handleWalletVerify(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var req struct {
		Message   string `json:"message" binding:"required"`
		Signature string `json:"signature" binding:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Extract address from message
	address, err := siwe.ExtractAddressFromMessage(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message format"})
		return
	}
	address = strings.ToLower(address)

	// Extract nonce from message
	nonce, err := siwe.ExtractNonceFromMessage(req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message format"})
		return
	}

	// Verify signature
	recoveredAddress, err := a.verifier.VerifySignature(req.Message, req.Signature)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Check if recovered address matches
	if strings.ToLower(recoveredAddress) != address {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "address mismatch"})
		return
	}

	// Get user
	user, err := a.storage.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify nonce matches
	if user.WalletVerificationNonce == nil || *user.WalletVerificationNonce != nonce {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid nonce"})
		return
	}

	// Link wallet to user
	if err := a.storage.LinkWalletToUser(c.Request.Context(), userID, address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to link wallet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet linked successfully"})
}

// handleEmailAdd initiates adding an email to a wallet user
func (a *API) handleEmailAdd(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Get user
	user, err := a.storage.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	// Check if email is already linked to another user
	existingUser, err := a.storage.GetUserByEmail(c.Request.Context(), req.Email)
	if err == nil && existingUser.ID != userID {
		c.JSON(http.StatusConflict, gin.H{"error": "email already linked to another account"})
		return
	}

	// If email is already linked to this user and verified, return success
	if user.Email != nil && strings.EqualFold(*user.Email, req.Email) {
		if user.EmailVerified {
			response := gin.H{
				"message":        "Email is already verified and linked to your account.",
				"email_verified": true,
			}
			c.JSON(http.StatusOK, response)
			return
		}
		// Email exists but not verified - we'll resend verification below
	}

	// Generate verification token
	token, err := email.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	expires := email.GetVerificationExpiry()

	// Store verification token (temporarily, will be linked after verification)
	// We'll use a special approach: create a temporary user or update existing
	// For now, we'll update the user's email field but mark as unverified
	// Then when they verify, we'll mark it as verified
	// Actually, we need to store the token first, then verify will link it

	// Update email verification fields
	if err := a.storage.UpdateEmailVerification(c.Request.Context(), userID, token, expires); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store verification token"})
		return
	}

	// Set email as unverified (or update if already exists)
	if user.Email == nil || !strings.EqualFold(*user.Email, req.Email) {
		// Only set email if it's new or different
		if err := a.storage.SetEmailUnverified(c.Request.Context(), userID, req.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set email"})
			return
		}
	}

	// Send verification email
	emailSent := false
	if a.emailService != nil {
		if err := a.emailService.SendVerificationEmail(req.Email, token); err != nil {
			log.Printf("Failed to send verification email: %v", err)
		} else {
			emailSent = true
		}
	}

	// Determine message based on whether email was already linked
	message := "Verification email sent. Please check your inbox."
	if user.Email != nil && strings.EqualFold(*user.Email, req.Email) && !user.EmailVerified {
		message = "Verification email resent. Please check your inbox to verify your email."
	}

	response := gin.H{
		"message":               message,
		"email_sent":            emailSent,
		"email_service_enabled": a.emailService != nil,
	}

	// If email service is disabled, include the token so user can verify manually
	if a.emailService == nil {
		verificationURL := fmt.Sprintf("%s/email-verify?token=%s", os.Getenv("BASE_URL"), token)
		if verificationURL == "/email-verify?token="+token {
			// Fallback if BASE_URL not set
			verificationURL = fmt.Sprintf("http://localhost:3000/email-verify?token=%s", token)
		}
		response["verification_token"] = token
		response["verification_url"] = verificationURL
		response["message"] = "Email service is not configured. Use the verification link below to verify your email."
		log.Printf("Email service disabled - Verification token for %s: %s", req.Email, token)
	}

	c.JSON(http.StatusOK, response)
}

// Run starts the HTTP server
func (a *API) Run(addr string) error {
	return a.router.Run(addr)
}
