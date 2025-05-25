package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/zahartd/social-network/src/services/stats-service/internal/config"
)

func InitClickhouse(cfg *config.Config) (clickhouse.Conn, error) {
	opts, err := clickhouse.ParseDSN(cfg.ClickhouseDSN)
	if err != nil {
		return nil, err
	}
	conn, err := clickhouse.Open(opts)
	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	log.Println("Connected to ClickHouse")
	return conn, nil
}
