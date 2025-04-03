package main

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/api-gateway/client"
	"github.com/zahartd/social-network/src/services/api-gateway/handlers"
)

var (
	rsaPublicKey *rsa.PublicKey
	postClient   postpb.PostServiceClient
)

type UserClaims struct {
	UserID string `json:"-"`
	jwt.RegisteredClaims
}

func loadRSAPublicKey() {
	pubKeyPath := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyPath == "" {
		log.Fatal("JWT_PUBLIC_KEY environment variable is not provided")
	}
	log.Printf("Loading public key from: %s", pubKeyPath)
	pubBytes, err := os.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatalf("Error reading RSA public key: %v", err)
	}
	rsaPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		log.Fatalf("Error parsing RSA public key: %v", err)
	}
	log.Println("RSA public key loaded successfully")
}

func validateJWT(tokenStr string) (*UserClaims, error) {
	tokenStr = strings.TrimSpace(strings.TrimPrefix(tokenStr, "Bearer"))
	if tokenStr == "" {
		return nil, errors.New("token string is empty")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			log.Printf("validateJWT KeyFunc: Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaPublicKey, nil
	})

	if err != nil {
		log.Printf("validateJWT: Token parsing or validation error: %v", err)
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
			log.Println("validateJWT: Subject ('sub') claim is missing or empty in token")
			return nil, errors.New("missing or empty Subject claim")
		}
		claims.UserID = userID

		log.Printf("validateJWT: Token validated successfully for Subject (UserID): %s", userID)
		return claims, nil
	}

	log.Println("validateJWT: Token is invalid after parsing")
	return nil, jwt.ErrTokenInvalidClaims
}

func proxyHandler(target *url.URL) gin.HandlerFunc {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		log.Printf("Proxying request to %s: %s %s", target, c.Request.Method, c.Request.URL.Path)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization token"})
			return
		}

		claims, err := validateJWT(tokenStr)
		if err != nil {
			log.Printf("Auth Middleware: Token validation failed: %v", err)
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token not yet valid"})
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token signature"})
			} else if err.Error() == "missing or empty Subject claim" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims (Subject missing)"})
			} else if strings.Contains(err.Error(), "unexpected signing method") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token (signing method)"})
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			}
			return
		}

		log.Printf("Auth Middleware: Auth successful for UserID: %s", claims.UserID)
		c.Set("userID", claims.UserID)

		c.Next()
	}
}

func main() {
	log.Println("Starting API Gateway...")
	loadRSAPublicKey()
	postClient = client.InitPostServiceClient()

	userServiceURLStr := os.Getenv("USER_SERVICE_URL")
	if userServiceURLStr == "" {
		log.Fatal("USER_SERVICE_URL environment variable is not provided")
	}
	userServiceURL, err := url.Parse(userServiceURLStr)
	if err != nil {
		log.Fatalf("Invalid USER_SERVICE_URL: %v", err)
	}
	log.Printf("User service proxy target URL: %s", userServiceURLStr)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// User Service Routes
	router.POST("/user", proxyHandler(userServiceURL))
	router.GET("/user/login", proxyHandler(userServiceURL))

	userProtected := router.Group("/user")
	userProtected.Use(authMiddleware())
	{
		userProtected.GET("/logout", proxyHandler(userServiceURL))
		userProtected.GET("/:identifier", proxyHandler(userServiceURL))
		userProtected.PUT("/:identifier", proxyHandler(userServiceURL))
		userProtected.DELETE("/:identifier", proxyHandler(userServiceURL))
	}
	log.Println("User service routes configured")

	postHandlers := handlers.NewPostHandler(postClient)

	router.GET("/posts", postHandlers.ListPublicPosts)
	log.Println("Public post listing route configured: GET /posts")

	postProtected := router.Group("/posts")
	postProtected.Use(authMiddleware())
	{
		postProtected.POST("", postHandlers.CreatePost)
		postProtected.GET("/my", postHandlers.ListMyPosts)
		postProtected.GET("/:postID", postHandlers.GetPost)
		postProtected.PUT("/:postID", postHandlers.UpdatePost)
		postProtected.DELETE("/:postID", postHandlers.DeletePost)
	}
	log.Println("Authenticated post service routes configured")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway listening on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
