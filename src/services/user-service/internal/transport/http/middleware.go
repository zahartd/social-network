package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/auth"
)

// AuthMiddleware attaches userID into context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := auth.TrimBearer(c.GetHeader("Authorization"))
		if tok == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := auth.ParseToken(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set("userID", claims["sub"])
		c.Set("token", tok)
		c.Next()
	}
}
