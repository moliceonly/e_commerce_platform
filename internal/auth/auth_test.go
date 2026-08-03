package auth

import (
	"testing"
	"time"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("123456")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" || hash == "123456" {
		t.Fatal("hash should not be empty or plain")
	}
	if !CheckPassword(hash, "123456") {
		t.Fatal("correct password should pass")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("wrong password should fail")
	}
}

func TestSignAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := SignToken(secret, 42, "user", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.Role != "user" {
		t.Fatalf("got uid=%d role=%s", claims.UserID, claims.Role)
	}
}

func TestParseToken_badSecret(t *testing.T) {
	token, err := SignToken("secret-a", 1, "user", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken("secret-b", token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestParseToken_expired(t *testing.T) {
	token, err := SignToken("test-secret", 1, "user", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken("test-secret", token); err == nil {
		t.Fatal("expected error for expired token")
	}
}
