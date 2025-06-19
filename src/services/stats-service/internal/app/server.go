package app

import (
	"net"

	"github.com/ClickHouse/clickhouse-go/v2"
	"google.golang.org/grpc"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
	"github.com/zahartd/social-network/src/services/stats-service/internal/config"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/service"
	clickhouse_repo "github.com/zahartd/social-network/src/services/stats-service/internal/infrastructure/clickhouse"
	kafka_consumer "github.com/zahartd/social-network/src/services/stats-service/internal/infrastructure/kafka"
	grpc_handler "github.com/zahartd/social-network/src/services/stats-service/internal/transport/grpc"
)

func Build(cfg *config.Config, lis net.Listener) (*grpc.Server, error) {
	opts, err := clickhouse.ParseDSN(cfg.ClickhouseDSN)
	if err != nil {
		return nil, err
	}
	ch, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}

	repo := clickhouse_repo.New(ch)

	consumer := kafka_consumer.New(cfg.KafkaBrokerURL, repo)
	consumer.RunAsync()

	svc := service.New(repo)

	srv := grpc.NewServer()
	statspb.RegisterStatsServiceServer(srv, grpc_handler.New(svc))

	return srv, nil
}
