package auth

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/zahartd/social-network/src/services/user-service/internal/models"
	"github.com/zahartd/social-network/src/services/user-service/internal/repository"
)

const tokenExpireTime = 3 * time.Minute

var (
	rsaPrivateKey *rsa.PrivateKey
	rsaPublicKey  *rsa.PublicKey
	sessionRepo   repository.SessionRepository
)

func InitJWT() {
	privKeyPath := os.Getenv("JWT_PRIVATE_KEY")
	if privKeyPath == "" {
		log.Fatal("JWT_PRIVATE_KEY not provided")
	}
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPath == "" {
		log.Fatal("JWT_PUBLIC_KEY not provided")
	}

	privBytes, err := os.ReadFile(privKeyPath)
	if err != nil {
		log.Fatalf("Error reading RSA private key: %v", err)
	}
	rsaPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		log.Fatalf("Error parsing RSA private key: %v", err)
	}

	pubBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalf("Error reading RSA public key: %v", err)
	}
	rsaPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		log.Fatalf("Error parsing RSA public key: %v", err)
	}
}

func SetSessionRepo(repo repository.SessionRepository) {
	sessionRepo = repo
}

func GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"login": user.Login,
		"exp":   time.Now().Add(tokenExpireTime).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(rsaPrivateKey)
}

func CreateSession(user *models.User, token string, ipAddress string) error {
	if sessionRepo == nil {
		return fmt.Errorf("session repository is not set")
	}
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s", ipAddress)
	}
	session := &models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(tokenExpireTime),
		IPAddress: ip,
	}
	return sessionRepo.CreateSession(session)
}

func TrimBearerPrefix(token string) string {
	const prefix = "Bearer "
	if len(token) > len(prefix) && token[:len(prefix)] == prefix {
		return token[len(prefix):]
	}
	return token
}

func DeleteSession(token string) error {
	return sessionRepo.DeleteSessionByToken(token)
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			log.Println("Missing Authorization token")
			c.AbortWithStatusJSON(401, gin.H{"error": "Missing Authorization token"})
			return
		}
		tokenStr = TrimBearerPrefix(tokenStr)
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return rsaPublicKey, nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
			return
		}
		sess, err := sessionRepo.GetSessionByToken(tokenStr)
		if err != nil {
			log.Printf("Session not found: %v", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "Session not found"})
			return
		}
		if time.Now().After(sess.ExpiresAt) {
			log.Println("Session expired")
			c.AbortWithStatusJSON(401, gin.H{"error": "Session expired"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}
		c.Set("userID", claims["sub"])
		c.Next()
	}
}
