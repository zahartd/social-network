package router

import (
	"github.com/gin-gonic/gin"

	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/handlers"
	"github.com/zahartd/social-network/src/services/user-service/internal/service"
)

func SetupRouter(r *gin.Engine, userService service.UserService) {
	userHandler := handlers.NewUserHandler(userService)

	r.POST("/user", userHandler.CreateUser)
	r.GET("/user/login", userHandler.Login)
	r.GET("/user/logout", userHandler.Logout)

	protected := r.Group("/user")
	protected.Use(auth.JWTAuthMiddleware())
	protected.GET("/:identifier", userHandler.GetUser)
	protected.PUT("/:identifier", userHandler.UpdateUser)
	protected.DELETE("/:identifier", userHandler.DeleteUser)
}
