package secret

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"key-haven-back/config"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrEmptyToken   = errors.New("empty token provided")
	ErrInvalidInput = errors.New("invalid input parameters")

	// Add a singleton pattern for the symmetric key
	symmetricKey     paseto.V4SymmetricKey
	symmetricKeyOnce sync.Once
	keyFingerprint   string // Store a fingerprint of the key for debugging
)

type TokenClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Generate a key fingerprint (hash) for debugging purposes
func generateKeyFingerprint(key paseto.V4SymmetricKey) string {
	hash := sha256.Sum256(key.ExportBytes())
	return hex.EncodeToString(hash[:])[:8] // First 8 chars is enough for fingerprinting
}

// saveKeyToFile saves the key to a file for persistence between restarts
func saveKeyToFile(keyHex string) error {
	// Determine the key storage location
	keyDir := config.GetEnvOrDefault("KEY_STORAGE_DIR", "./keys")
	keyPath := filepath.Join(keyDir, "paseto.key")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write the key with restricted permissions
	return os.WriteFile(keyPath, []byte(keyHex), 0600)
}

// loadKeyFromFile attempts to load a previously saved key
func loadKeyFromFile() (string, error) {
	keyDir := config.GetEnvOrDefault("KEY_STORAGE_DIR", "./keys")
	keyPath := filepath.Join(keyDir, "paseto.key")

	data, err := os.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Generate a secure key or load from environment
func getOrCreateSymmetricKey() paseto.V4SymmetricKey {
	// Use sync.Once to ensure the key is only created once per application lifetime
	symmetricKeyOnce.Do(func() {
		// First, check environment variable
		keyStr := config.GetEnvOrDefault("PASETO_KEY", "")

		// If we have a key from the environment, use it
		if keyStr != "" {
			log.Printf("Using PASETO key from environment variable")

			var keyBytes []byte
			var err error

			// Try to decode as base64 first (with and without padding)
			keyBytes, err = base64.StdEncoding.DecodeString(keyStr)
			if err != nil {
				// Try without padding
				keyBytes, err = base64.RawStdEncoding.DecodeString(keyStr)
			}

			// If base64 fails, try hex decoding
			if err != nil {
				keyBytes, err = hex.DecodeString(keyStr)
				if err != nil {
					log.Printf("Error decoding PASETO key from environment: %v. This is critical - key must be valid base64 or hex.", err)
					// We don't fall back to a new key here because that would defeat the purpose of setting the env var
					panic("PASETO_KEY environment variable is not valid base64 or hexadecimal")
				}
			}

			// Use the correct function to load a symmetric key
			key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
			if err != nil {
				log.Printf("Error creating symmetric key from environment: %v. This is critical.", err)
				panic("Failed to create PASETO key from environment variable")
			}

			symmetricKey = key
			keyFingerprint = generateKeyFingerprint(symmetricKey)
			log.Printf("Successfully loaded PASETO key from environment (Fingerprint: %s)", keyFingerprint)
			return
		}

		// If not in environment, try to load from file
		loadedKeyStr, err := loadKeyFromFile()
		if err == nil {
			// Try to decode the key from file - same logic as above
			var keyBytes []byte
			var decodeErr error

			// Try to decode as base64 first
			keyBytes, decodeErr = base64.StdEncoding.DecodeString(loadedKeyStr)
			if decodeErr != nil {
				// Try without padding
				keyBytes, decodeErr = base64.RawStdEncoding.DecodeString(loadedKeyStr)
				if decodeErr != nil {
					// Finally try hex
					keyBytes, decodeErr = hex.DecodeString(loadedKeyStr)
				}
			}

			if decodeErr != nil {
				log.Printf("Error decoding PASETO key from file: %v.", decodeErr)
			} else {
				key, err := paseto.V4SymmetricKeyFromBytes(keyBytes)
				if err == nil {
					symmetricKey = key
					keyFingerprint = generateKeyFingerprint(symmetricKey)
					log.Printf("Using PASETO key from file storage (Fingerprint: %s)", keyFingerprint)
					return
				}
			}
		}

		// If we get here, we need to generate a new key
		symmetricKey = paseto.NewV4SymmetricKey()
		// Use base64 encoding for the key going forward
		keyBase64Str := base64.StdEncoding.EncodeToString(symmetricKey.ExportBytes())

		// Save the key for future use
		if err := saveKeyToFile(keyBase64Str); err != nil {
			log.Printf("Warning: Failed to save PASETO key: %v", err)
		}

		keyFingerprint = generateKeyFingerprint(symmetricKey)
		log.Printf("IMPORTANT: Generated new PASETO key. Add this to your .env file:")
		log.Printf("PASETO_KEY=%s", keyBase64Str)
	})

	return symmetricKey
}

// GenerateToken creates a secure PASETO token with user information
func GenerateToken(userID, email string, expiresIn time.Duration) (string, error) {
	// Validate input
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(email) == "" {
		return "", fmt.Errorf("%w: userID and email cannot be empty", ErrInvalidInput)
	}

	if expiresIn <= 0 {
		return "", fmt.Errorf("%w: expiration time must be positive", ErrInvalidInput)
	}

	key := getOrCreateSymmetricKey()
	log.Printf("Generating token with key fingerprint: %s", keyFingerprint)
	token := paseto.NewToken()

	// Add standard claims
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(expiresIn))

	// Add custom claims
	token.SetString("user_id", userID)
	token.SetString("email", email)

	// Generate a random nonce for added security
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	token.SetString("nonce", hex.EncodeToString(nonce))

	// Encrypt and sign the token
	encrypted := token.V4Encrypt(key, nil)
	return encrypted, nil
}

// ValidateToken verifies and extracts claims from a PASETO token
func ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, ErrEmptyToken
	}

	// Clean the token - remove any unwanted characters that might affect decoding
	tokenString = strings.TrimSpace(tokenString)

	// Check if token matches expected format for PASETO v4.local tokens
	if !strings.HasPrefix(tokenString, "v4.local.") {
		return nil, fmt.Errorf("%w: token doesn't have valid PASETO v4.local prefix", ErrInvalidToken)
	}

	// Log token info for debugging
	log.Printf("Processing token of length: %d", len(tokenString))
	log.Printf("Validating token with key fingerprint: %s", keyFingerprint)

	// Check for common issues that might cause Base64 corruption
	if strings.Contains(tokenString, " ") || strings.Contains(tokenString, "\n") || strings.Contains(tokenString, "\t") {
		return nil, fmt.Errorf("%w: token contains whitespace characters", ErrInvalidToken)
	}

	// Get the singleton key
	key := getOrCreateSymmetricKey()
	parser := paseto.NewParser()

	// Set validation rules with more lenient time settings
	parser.AddRule(paseto.NotExpired())
	// Use a time skew allowance to handle small clock differences
	//parser.AddRule(paseto.ValidAt(time.Now().Add(-2 * time.Minute))) // Allow for 2 minutes of clock skew

	// Try to parse the token with better error handling
	var token *paseto.Token
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%w: parser panic: %v", ErrInvalidToken, r)
				log.Printf("PASETO parser panic: %v", r)
			}
		}()

		// Try parsing with more detailed error logging
		token, err = parser.ParseV4Local(key, tokenString, nil)
		if err != nil {
			if strings.Contains(err.Error(), "message authentication code") {
				log.Printf("MAC verification failed: This typically means the token was created with a different key")
			} else {
				log.Printf("Token parsing failed: %v", err)
			}
		}
	}()

	// Handle parsing errors
	if err != nil {
		if strings.Contains(err.Error(), "token has expired") {
			return nil, ErrTokenExpired
		}

		// Special handling for MAC errors which indicate key mismatch
		if strings.Contains(err.Error(), "message authentication code") {
			return nil, fmt.Errorf("%w: token was likely created with a different key (key fingerprint: %s)",
				ErrInvalidToken, keyFingerprint)
		}

		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Extract claims
	userID, err := token.GetString("user_id")
	if err != nil {
		return nil, fmt.Errorf("%w: missing user_id claim", ErrInvalidToken)
	}

	email, err := token.GetString("email")
	if err != nil {
		return nil, fmt.Errorf("%w: missing email claim", ErrInvalidToken)
	}

	// Extract expiration time
	expClaim, err := token.GetTime("exp")
	if err != nil {
		return nil, fmt.Errorf("%w: missing expiration claim", ErrInvalidToken)
	}

	log.Printf("Successfully validated token for user %s", email)

	return &TokenClaims{
		UserID:    userID,
		Email:     email,
		ExpiresAt: expClaim,
	}, nil
}

// RefreshToken creates a new token with fresh expiration time
func RefreshToken(tokenString string, newExpiresIn time.Duration) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate a new token with the same user data
	return GenerateToken(claims.UserID, claims.Email, newExpiresIn)
}
