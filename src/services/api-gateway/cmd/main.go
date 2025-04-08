package main

import (
	"log"
	"net/url"
	"os"

	"github.com/zahartd/social-network/src/services/api-gateway/internal/auth"
	"github.com/zahartd/social-network/src/services/api-gateway/internal/client"
	"github.com/zahartd/social-network/src/services/api-gateway/internal/router"
)

func main() {
	if err := auth.LoadRSAPublicKey(); err != nil {
		log.Fatalf("Failed to load RSA public key: %v", err)
	}

	postClient := client.InitPostServiceClient()

	userServiceURLStr := os.Getenv("USER_SERVICE_URL")
	if userServiceURLStr == "" {
		log.Fatal("USER_SERVICE_URL environment variable is not provided")
	}
	userServiceURL, err := url.Parse(userServiceURLStr)
	if err != nil {
		log.Fatalf("Invalid USER_SERVICE_URL: %v", err)
	}

	r := router.SetupRouter(postClient, userServiceURL)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
