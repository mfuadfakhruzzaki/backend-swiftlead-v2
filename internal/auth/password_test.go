package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "securepassword"
	hash, err := HashPassword(password)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not match plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "securepassword"
	hash, _ := HashPassword(password)

	if !CheckPassword(hash, password) {
		t.Error("Password check failed for correct password")
	}

	if CheckPassword(hash, "wrongpassword") {
		t.Error("Password check passed for wrong password")
	}
}
