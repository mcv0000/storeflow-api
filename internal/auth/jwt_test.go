package auth

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	userID := "test-user-id-123"
	secret := "test-secret-key"

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	userID := "test-user-id-123"
	secret := "test-secret-key"

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("Expected userID %s, got %s", userID, claims.UserID)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	userID := "test-user-id-123"
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"

	token, err := GenerateToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, wrongSecret)
	if err == nil {
		t.Fatal("ValidateToken should fail with wrong secret")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	invalidToken := "invalid.token.here"

	_, err := ValidateToken(invalidToken, secret)
	if err == nil {
		t.Fatal("ValidateToken should fail with invalid token")
	}
}
