package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization token"})
			return
		}

		claims, err := validateJWT(tokenStr)
		if err != nil {
			var httpStatus int = http.StatusUnauthorized
			var errMsg string

			if errors.Is(err, jwt.ErrTokenExpired) {
				errMsg = "Token has expired"
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				errMsg = "Token not yet valid"
			} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				errMsg = "Invalid token signature"
			} else if err.Error() == "missing or empty Subject claim in token" {
				errMsg = "Invalid token claims (Subject missing)"
			} else if strings.Contains(err.Error(), "unexpected signing method") {
				errMsg = "Invalid token (signing method)"
			} else if err.Error() == "token string is empty after trimming Bearer prefix" {
				errMsg = "Invalid Authorization header format"
			} else {
				errMsg = "Invalid token"
			}

			c.AbortWithStatusJSON(httpStatus, gin.H{"error": errMsg})
			return
		}

		c.Set("userID", claims.UserID)

		c.Next()
	}
}
