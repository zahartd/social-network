package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/handlers"
	"github.com/zahartd/social-network/src/services/user-service/internal/repository"
	"github.com/zahartd/social-network/src/services/user-service/internal/service"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not provided")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepo(db)
	sessionRepo := repository.NewPostgresSessionRepo(db)
	auth.SetSessionRepo(sessionRepo)
	userService := service.NewUserService(userRepo, sessionRepo)
	userHandler := handlers.NewUserHandler(userService)

	auth.InitJWT()

	router := gin.Default()

	router.POST("/user", userHandler.CreateUser)
	router.GET("/user/login", userHandler.Login)
	router.GET("/user/logout", userHandler.Logout)

	protected := router.Group("/user")
	protected.Use(auth.JWTAuthMiddleware())
	protected.GET("/:identifier", userHandler.GetUser)
	protected.PUT("/:identifier", userHandler.UpdateUser)
	protected.DELETE("/:identifier", userHandler.DeleteUser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("User Service running on port %s", port)
	router.Run(":" + port)
}
