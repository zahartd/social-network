package config

import (
	"log"
	"os"
)

type Config struct {
	Port           string
	DB_DSN         string
	KafkaBrokerURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	return &Config{
		Port:           port,
		DB_DSN:         dbDSN,
		KafkaBrokerURL: os.Getenv("KAFKA_BROKER_URL"),
	}
}
