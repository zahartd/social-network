package config

import (
	"log"
	"os"
)

type Config struct {
	GRPCPort       string
	ClickhouseDSN  string
	KafkaBrokerURL string
}

func Load() *Config {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	ClickhouseDSN := os.Getenv("CLICKHOUSE_DSN")
	if ClickhouseDSN == "" {
		log.Fatal("DB_DSN environment variable is not set")
	}

	return &Config{
		GRPCPort:       port,
		ClickhouseDSN:  ClickhouseDSN,
		KafkaBrokerURL: os.Getenv("KAFKA_BROKER_URL"),
	}
}
