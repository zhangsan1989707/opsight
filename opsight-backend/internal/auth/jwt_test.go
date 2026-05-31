package auth

import (
	"testing"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-unit-tests")

	token, err := GenerateToken(42, "test@example.com", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected user_id 42, got %d", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", claims.Email)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-unit-tests")

	_, err := ValidateToken("not.a.valid.token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret-a")
	token, err := GenerateToken(1, "a@b.com", "viewer")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	t.Setenv("JWT_SECRET", "secret-b")
	_, err = ValidateToken(token)
	if err == nil {
		t.Error("expected error validating token signed with different secret")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-for-unit-tests")
	t.Setenv("JWT_EXPIRY_HOURS", "1")

	token, err := GenerateToken(1, "a@b.com", "viewer")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Token should be valid immediately
	_, err = ValidateToken(token)
	if err != nil {
		t.Errorf("token should be valid immediately after creation: %v", err)
	}
}
