package main

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateJWTValidToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	rsaPublicKey = &privKey.PublicKey
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":   "123",
		"login": "testuser",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	})
	tokenStr, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	tokenStrWithBearer := "Bearer " + tokenStr
	if err := validateJWT(tokenStrWithBearer); err != nil {
		t.Errorf("Expected valid token, got error: %v", err)
	}
}

func TestValidateJWTInvalidToken(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}
	rsaPublicKey = &privKey.PublicKey
	invalidToken := "Bearer invalid.token.value"
	if err := validateJWT(invalidToken); err == nil {
		t.Error("Expected error for invalid token, but got nil")
	}
}
