package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zahartd/social-network/src/services/user-service/internal/models"
)

func TestGenerateToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	rsaPrivateKey = privKey
	rsaPublicKey = &privKey.PublicKey
	user := &models.User{
		ID:    "test-uuid",
		Login: "testuser",
	}

	tokenStr, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return rsaPublicKey, nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}
	if !parsedToken.Valid {
		t.Fatal("Token is not valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to cast claims to jwt.MapClaims")
	}
	if claims["sub"] != user.ID {
		t.Errorf("Expected sub claim %s, got %v", user.ID, claims["sub"])
	}
	if claims["login"] != user.Login {
		t.Errorf("Expected login claim %s, got %v", user.Login, claims["login"])
	}

	iat, ok1 := claims["iat"].(float64)
	exp, ok2 := claims["exp"].(float64)
	if !ok1 || !ok2 {
		t.Error("iat or exp claim has invalid type")
	} else if exp <= iat {
		t.Errorf("Expected exp > iat, got iat=%v and exp=%v", iat, exp)
	}
}

func TestTrimBearerPrefix(t *testing.T) {
	tokenWithPrefix := "Bearer mytoken225"
	expected := "mytoken225"
	result := TrimBearerPrefix(tokenWithPrefix)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	tokenWithoutPrefix := "mytoken225"
	result = TrimBearerPrefix(tokenWithoutPrefix)
	if result != tokenWithoutPrefix {
		t.Errorf("Expected %q, got %q", tokenWithoutPrefix, result)
	}

	result = TrimBearerPrefix("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}
