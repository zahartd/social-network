package auth

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/zahartd/social-network/src/services/user-service/internal/domain/models"
)

const TokenTTL = 3 * time.Minute

var (
	priv *rsa.PrivateKey
	pub  *rsa.PublicKey
)

func Init() {
	privPath := os.Getenv("JWT_PRIVATE_KEY")
	pubPath := os.Getenv("JWT_PUBLIC_KEY")
	if privPath == "" || pubPath == "" {
		panic("JWT_PRIVATE_KEY / JWT_PUBLIC_KEY not set")
	}
	pBytes, _ := os.ReadFile(privPath)
	priv, _ = jwt.ParseRSAPrivateKeyFromPEM(pBytes)
	qBytes, _ := os.ReadFile(pubPath)
	pub, _ = jwt.ParseRSAPublicKeyFromPEM(qBytes)
}

func GenerateToken(u *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   u.ID,
		"login": u.Login,
		"exp":   time.Now().Add(TokenTTL).Unix(),
		"iat":   time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return t.SignedString(priv)
}

func ParseToken(tokenStr string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pub, nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return c, nil
}

const bearer = "Bearer "

func TrimBearer(s string) string {
	if len(s) > len(bearer) && s[:len(bearer)] == bearer {
		return s[len(bearer):]
	}
	return s
}
