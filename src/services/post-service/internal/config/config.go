package config

import (
	"log"
	"os"
)

type Config struct {
	GRPCPort string
	DB_DSN   string
}

func Load() *Config {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	return &Config{
		GRPCPort: port,
		DB_DSN:   dbDSN,
	}
}
