package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

func prepareKeys(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("cannot generate RSA key: %v", err)
	}
	priv = key
	pub = &key.PublicKey
}

func TestGenerateAndParseToken(t *testing.T) {
	prepareKeys(t)

	user := &models.User{ID: "uuid-123", Login: "testuser"}

	tokenStr, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims["sub"] != user.ID {
		t.Errorf("sub claim mismatch: want %s, got %v", user.ID, claims["sub"])
	}
	if claims["login"] != user.Login {
		t.Errorf("login claim mismatch: want %s, got %v", user.Login, claims["login"])
	}

	iat, ok1 := claims["iat"].(float64)
	exp, ok2 := claims["exp"].(float64)
	if !ok1 || !ok2 {
		t.Fatal("iat/exp claims have wrong type")
	}
	if exp <= iat {
		t.Errorf("exp (%v) should be greater than iat (%v)", exp, iat)
	}

	ttl := time.Duration(int64(exp-iat)) * time.Second
	if ttl != TokenTTL {
		t.Logf("warn: unexpected TTL: got %v, want %v", ttl, TokenTTL)
	}
}

func TestTrimBearer(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Bearer abc123", "abc123"},
		{"abc123", "abc123"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := TrimBearer(tc.in); got != tc.want {
			t.Errorf("TrimBearer(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
