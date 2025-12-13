package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcryptCost is the computational cost for bcrypt hashing (2^12 iterations)
	// Higher cost = more secure but slower. 12 is a good balance for 2024.
	bcryptCost = 12

	// minPasswordLength is the minimum recommended password length
	minPasswordLength = 8

	// defaultTokenLength is the default length for random tokens in bytes
	defaultTokenLength = 32
)

// HashPassword hashes a password using bcrypt with cost factor 12
//
// Bcrypt automatically handles:
// - Salt generation (random, unique per password)
// - Multiple rounds of hashing (2^12 = 4096 iterations)
// - Constant-time comparison protection
//
// The returned hash includes the algorithm, cost, salt, and hash in a single string.
// Format: $2a$12$[22 char salt][31 char hash]
//
// Example:
//
//	hash, err := HashPassword("MySecurePass123")
//	if err != nil {
//	    return err
//	}
//	// Store hash in database
//	user.PasswordHash = hash
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) < minPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// CheckPassword compares a plaintext password with a bcrypt hash
//
// Uses constant-time comparison to prevent timing attacks.
// Returns true if password matches hash, false otherwise.
//
// This function is safe to use in authentication flows.
//
// Example:
//
//	valid := CheckPassword(inputPassword, user.PasswordHash)
//	if !valid {
//	    return ErrInvalidCredentials
//	}
func CheckPassword(password, hash string) bool {
	if password == "" || hash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomToken generates a cryptographically secure random token
//
// The token is hex-encoded, so the output length will be length*2 characters.
// For example, length=32 produces a 64-character hex string.
//
// Use cases:
// - Password reset tokens
// - Email verification tokens
// - API keys
// - Session tokens
//
// Example:
//
//	token, err := GenerateRandomToken(32)
//	if err != nil {
//	    return err
//	}
//	// Store hash of token in database
//	user.ResetToken = HashToken(token)
//	// Send plain token to user (only once, via email)
//	sendEmail(user.Email, "Reset link: /reset?token=" + token)
func GenerateRandomToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}

	if length > 1024 {
		return "", fmt.Errorf("token length cannot exceed 1024 bytes")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// GenerateToken is a convenience wrapper for GenerateRandomToken with default length
//
// Generates a 32-byte (64 hex char) token suitable for most use cases.
//
// Example:
//
//	token, err := GenerateToken()
//	// Returns 64-character hex string
func GenerateToken() (string, error) {
	return GenerateRandomToken(defaultTokenLength)
}

// HashToken creates a SHA-256 hash of a token for secure storage
//
// IMPORTANT: Store only the hash in your database, never the plain token.
// The plain token should only be shown to the user once (e.g., in email).
//
// SHA-256 is one-way: you cannot recover the original token from the hash.
// To verify a token, hash the user-provided token and compare hashes.
//
// Example:
//
//	// Generate and store token
//	plainToken, _ := GenerateToken()
//	hashedToken := HashToken(plainToken)
//	db.SaveResetToken(userID, hashedToken)
//	sendEmail(user.Email, plainToken)
//
//	// Later, verify token from URL
//	providedToken := r.URL.Query().Get("token")
//	providedHash := HashToken(providedToken)
//	if providedHash == user.ResetTokenHash {
//	    // Valid token
//	}
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateHMAC creates an HMAC-SHA256 signature for webhook/API verification
//
// HMAC (Hash-based Message Authentication Code) ensures:
// - Message authenticity (sender has the secret key)
// - Message integrity (content hasn't been tampered with)
//
// Common use cases:
// - Webhook signatures (GitHub, Stripe, etc.)
// - API request signing
// - JWT signatures (though use proper JWT libraries for that)
//
// Example:
//
//	secret := "webhook-secret-key"
//	payload := `{"event": "user.created", "user_id": 123}`
//	signature := GenerateHMAC(secret, payload)
//
//	// Include signature in webhook header
//	req.Header.Set("X-Signature", signature)
func GenerateHMAC(secret, payload string) string {
	if secret == "" || payload == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	signature := h.Sum(nil)

	return hex.EncodeToString(signature)
}

// VerifyHMAC verifies an HMAC-SHA256 signature using constant-time comparison
//
// IMPORTANT: Uses hmac.Equal() for constant-time comparison to prevent timing attacks.
// Never use == or bytes.Equal() to compare signatures.
//
// Returns true if signature is valid, false otherwise.
//
// Example:
//
//	// Webhook receiver
//	receivedSig := r.Header.Get("X-Signature")
//	payload, _ := io.ReadAll(r.Body)
//	secret := os.Getenv("WEBHOOK_SECRET")
//
//	if !VerifyHMAC(secret, string(payload), receivedSig) {
//	    return errors.New("invalid webhook signature")
//	}
func VerifyHMAC(secret, payload, signature string) bool {
	if secret == "" || payload == "" || signature == "" {
		return false
	}

	expectedSignature := GenerateHMAC(secret, payload)

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// SecureCompare performs constant-time comparison of two strings
//
// Use this instead of == when comparing:
// - Passwords
// - Tokens
// - API keys
// - Any security-sensitive values
//
// Prevents timing attacks where an attacker can deduce information
// by measuring how long comparisons take.
//
// Example:
//
//	if !SecureCompare(providedToken, storedToken) {
//	    return ErrUnauthorized
//	}
func SecureCompare(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}
