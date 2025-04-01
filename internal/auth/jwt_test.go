package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	// Setup
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 24 * time.Hour

	// Test successful JWT creation
	t.Run("Success", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)

		// Assert no errors
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Error("Expected non-empty token")
		}

		// Verify token can be parsed
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

		if err != nil {
			t.Errorf("Failed to parse token: %v", err)
		}
		if !parsedToken.Valid {
			t.Error("Expected token to be valid")
		}

		// Verify claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			t.Error("Failed to extract claims")
		}
		if claims["iss"] != "chirpy" {
			t.Errorf("Expected issuer to be 'chirpy', got %v", claims["iss"])
		}
		if claims["sub"] != userID.String() {
			t.Errorf("Expected subject to be '%s', got %v", userID.String(), claims["sub"])
		}

		// Verify expiration
		expiryTime := time.Unix(int64(claims["exp"].(float64)), 0)
		expectedExpiry := time.Now().Add(expiresIn)
		timeDiff := expiryTime.Sub(expectedExpiry)
		if timeDiff < -5*time.Second || timeDiff > 5*time.Second {
			t.Errorf("Expected expiry time around %v, got %v", expectedExpiry, expiryTime)
		}
	})

	// Test with empty secret
	t.Run("EmptySecret", func(t *testing.T) {
		token, err := MakeJWT(userID, "", expiresIn)
		if err == nil {
			t.Error("Expected an error with empty secret, got none")
		}
		if token != "" {
			t.Errorf("Expected empty token, got %s", token)
		}
	})
}

func TestValidateJWT(t *testing.T) {
	// Setup
	userID := uuid.New()
	tokenSecret := "test-secret"
	expiresIn := 24 * time.Hour

	// Test successful validation
	t.Run("Success", func(t *testing.T) {
		// Create a token first
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Validate the token
		validatedUserID, err := ValidateJWT(token, tokenSecret)

		// Assert validation success
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if validatedUserID != userID {
			t.Errorf("Expected user ID %v, got %v", userID, validatedUserID)
		}
	})

	// Test with wrong secret
	t.Run("WrongSecret", func(t *testing.T) {
		token, err := MakeJWT(userID, tokenSecret, expiresIn)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		validatedUserID, err := ValidateJWT(token, "wrong-secret")
		if err == nil {
			t.Error("Expected an error with wrong secret, got none")
		}
		if validatedUserID != uuid.Nil {
			t.Errorf("Expected nil UUID, got %v", validatedUserID)
		}
	})

	// Test with expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		// Create a token that expires immediately
		token, err := MakeJWT(userID, tokenSecret, -1*time.Hour)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		validatedUserID, err := ValidateJWT(token, tokenSecret)
		if err == nil {
			t.Error("Expected an error with expired token, got none")
		}
		if validatedUserID != uuid.Nil {
			t.Errorf("Expected nil UUID, got %v", validatedUserID)
		}
	})

	// Test with invalid token format
	t.Run("InvalidTokenFormat", func(t *testing.T) {
		validatedUserID, err := ValidateJWT("invalid-token", tokenSecret)
		if err == nil {
			t.Error("Expected an error with invalid token format, got none")
		}
		if validatedUserID != uuid.Nil {
			t.Errorf("Expected nil UUID, got %v", validatedUserID)
		}
	})

	// Test with invalid UUID in subject
	t.Run("InvalidUUID", func(t *testing.T) {
		// Create a custom token with an invalid UUID
		claims := &jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   "not-a-uuid",
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(tokenSecret))
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		validatedUserID, err := ValidateJWT(tokenString, tokenSecret)
		if err == nil {
			t.Error("Expected an error with invalid UUID, got none")
		}
		if validatedUserID != uuid.Nil {
			t.Errorf("Expected nil UUID, got %v", validatedUserID)
		}
	})
}
