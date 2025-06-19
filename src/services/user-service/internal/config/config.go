package config

import "github.com/caarlos0/env/v10"

type Config struct {
	HTTP struct {
		Port string `env:"PORT" envDefault:"8081"`
	}
	DB struct {
		DSN string `env:"DB_DSN,required"`
	}
	Kafka struct {
		BrokerURL string `env:"KAFKA_BROKER_URL"`
	}
}

func MustLoad() *Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return &cfg
}
