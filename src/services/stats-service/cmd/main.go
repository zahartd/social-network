package main

import (
	"log"
	"net"

	"github.com/zahartd/social-network/src/services/stats-service/internal/config"
	"github.com/zahartd/social-network/src/services/stats-service/internal/consumer"
	"github.com/zahartd/social-network/src/services/stats-service/internal/handlers"
	"github.com/zahartd/social-network/src/services/stats-service/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	statspb "github.com/zahartd/social-network/src/gen/go/stats"
)

func main() {
	cfg := config.Load()

	ck, err := storage.InitClickhouse(cfg)
	if err != nil {
		log.Fatalf("ClickHouse init failed: %v", err)
	}

	go consumer.RunAll(cfg, ck)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	handler := handlers.NewGRPCHandler(ck)
	statspb.RegisterStatsServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	log.Printf("Stats-service listening on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC serve error: %v", err)
	}
}
