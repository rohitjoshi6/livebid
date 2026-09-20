package auth

import "testing"

func TestPasswordHashDoesNotStorePlaintext(t *testing.T) {
	password := "correct-horse-battery"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == password {
		t.Fatal("password hash must not equal plaintext password")
	}
	if !CheckPassword(password, hash) {
		t.Fatal("expected password to verify against hash")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("wrong password should not verify")
	}
}
