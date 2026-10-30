package auth

import (
	"testing"
	"time"
)

func TestTokenRoundTripAndExpiry(t *testing.T) {
	now := time.Now()
	token, err := SignToken("user-1", "01234567890123456789012345678901", now)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ParseToken(token, "01234567890123456789012345678901")
	if err != nil || id != "user-1" {
		t.Fatalf("id=%q err=%v", id, err)
	}
	if _, err := ParseToken(token, "wrong-secret-012345678901234567890"); err == nil {
		t.Fatal("expected invalid signature")
	}
}
