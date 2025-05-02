package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"

	"github.com/zahartd/social-network/src/services/user-service/internal/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/config"
	"github.com/zahartd/social-network/src/services/user-service/internal/handlers"
	"github.com/zahartd/social-network/src/services/user-service/internal/repository"
	"github.com/zahartd/social-network/src/services/user-service/internal/service"
	"github.com/zahartd/social-network/src/services/user-service/internal/utils"
)

func PhoneValidator(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return utils.ValidatePhone(phone)
}

func PasswordValidator(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return utils.ValidatePassword(password)
}

func LoginValidator(fl validator.FieldLevel) bool {
	login, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return utils.ValidateLogin(login)
}

func initCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("phone", PhoneValidator); err != nil {
			panic(err)
		}
		if err := v.RegisterValidation("password", PasswordValidator); err != nil {
			panic(err)
		}
		if err := v.RegisterValidation("login", LoginValidator); err != nil {
			panic(err)
		}
	}
}

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepo(db)
	sessionRepo := repository.NewPostgresSessionRepo(db)
	auth.SetSessionRepo(sessionRepo)
	registrationsWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokerURL),
		Topic:                  "user-registrations",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	defer func() {
		if err := registrationsWriter.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()
	userService := service.NewUserService(userRepo, sessionRepo, registrationsWriter)
	userHandler := handlers.NewUserHandler(userService)

	auth.InitJWT()

	router := gin.Default()

	initCustomValidators()

	router.POST("/user", userHandler.CreateUser)
	router.GET("/user/login", userHandler.Login)
	router.GET("/user/logout", userHandler.Logout)

	protected := router.Group("/user")
	protected.Use(auth.JWTAuthMiddleware())
	protected.GET("/:identifier", userHandler.GetUser)
	protected.PUT("/:identifier", userHandler.UpdateUser)
	protected.DELETE("/:identifier", userHandler.DeleteUser)

	router.Run(":" + cfg.Port)
}
