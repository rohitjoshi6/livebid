package auth

import (
	"testing"
	"time"
)

func TestTokenManagerCreateAndParse(t *testing.T) {
	manager := NewTokenManager("test-secret", "livebid-test", time.Hour)
	token, err := manager.Create("user-1", "buyer@example.com", time.Now())
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	claims, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-1" || claims.Email != "buyer@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	manager := NewTokenManager("test-secret", "livebid-test", time.Hour)
	if _, err := manager.Parse("not-a-token"); err == nil {
		t.Fatal("expected invalid token to be rejected")
	}
}
