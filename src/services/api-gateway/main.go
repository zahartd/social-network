package main

import (
	"crypto/rsa"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var rsaPublicKey *rsa.PublicKey

func loadRSAPublicKey() {
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPath == "" {
		log.Fatal("JWT_PUBLIC_KEY is not provided")
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

func validateJWT(tokenStr string) error {
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "Bearer"))
	_, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrTokenInvalidIssuer
		}
		return rsaPublicKey, nil
	})
	return err
}

func proxyHandler(target *url.URL) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			log.Println("Missing Authorization token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization token"})
			return
		}
		if err := validateJWT(token); err != nil {
			log.Printf("Invalid token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}

func main() {
	loadRSAPublicKey()

	userServiceURLStr := os.Getenv("USER_SERVICE_URL")
	if userServiceURLStr == "" {
		log.Fatal("USER_SERVICE_URL is not provided")
	}
	userServiceURL, err := url.Parse(userServiceURLStr)
	if err != nil {
		log.Fatalf("Invalid USER_SERVICE_URL: %v", err)
	}

	router := gin.Default()

	router.POST("/user", proxyHandler(userServiceURL))
	router.GET("/user/login", proxyHandler(userServiceURL))
	router.GET("/user/logout", proxyHandler(userServiceURL))

	protected := router.Group("/user")
	protected.Use(authMiddleware())
	protected.GET("/:identifier", proxyHandler(userServiceURL))
	protected.PUT("/:identifier", proxyHandler(userServiceURL))
	protected.DELETE("/:identifier", proxyHandler(userServiceURL))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting API Gateway on port %s", port)
	router.Run(":" + port)
}
