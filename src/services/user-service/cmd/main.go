package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/zahartd/social-network/src/services/user-service/internal/app"
	"github.com/zahartd/social-network/src/services/user-service/internal/config"
	"github.com/zahartd/social-network/src/services/user-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/auth"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/kafka"
	"github.com/zahartd/social-network/src/services/user-service/internal/infrastructure/postgres"
	"github.com/zahartd/social-network/src/services/user-service/internal/transport/http"
)

func main() {
	cfg := config.MustLoad()

	// --- infra setup ----------------------------------------------------
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	userRepo := postgres.NewUserRepo(db)
	sessRepo := postgres.NewSessionRepo(db)

	writer := kafka.NewUserWriter(cfg.Kafka.BrokerURL)
	defer writer.Close()

	// JWT (private/public key paths are taken from env inside Init)
	auth.Init()

	// --- services -------------------------------------------------------
	userSvc := service.NewUser(userRepo, sessRepo, writer)

	// --- HTTP -----------------------------------------------------------
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	app.RegisterValidators(r)

	http.AttachRoutes(r, userSvc)

	srv := app.NewServer(r, cfg.HTTP.Port)

	log.Printf("⇢ user‑service started on :%s", cfg.HTTP.Port)
	if err := srv.ListenAndServe(); err != nil && err != context.Canceled {
		log.Fatal(err)
	}
}
