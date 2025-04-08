package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var rsaPublicKey *rsa.PublicKey

type UserClaims struct {
	UserID string `json:"-"`
	jwt.RegisteredClaims
}

func LoadRSAPublicKey() error {
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPath == "" {
		return errors.New("JWT_PUBLIC_KEY environment variable is not provided")
	}

	pubBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("error reading RSA public key file %s: %w", pubKeyPath, err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return fmt.Errorf("error parsing RSA public key from PEM: %w", err)
	}

	rsaPublicKey = key
	return nil
}

func validateJWT(tokenStr string) (*UserClaims, error) {
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "Bearer"))
	if tokenStr == "" {
		return nil, errors.New("token string is empty after trimming Bearer prefix")
	}

	if rsaPublicKey == nil {
		return nil, errors.New("authentication service is not properly configured (missing public key)")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, jwt.ErrTokenNotValidYet
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		userID := claims.RegisteredClaims.Subject
		if userID == "" {
			return nil, errors.New("missing or empty Subject claim in token")
		}
		claims.UserID = userID
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
