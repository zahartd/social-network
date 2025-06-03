package client

import (
	"log"
	"os"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitStatsServiceClient() statspb.StatsServiceClient {
	statsServiceURL := os.Getenv("STATS_SERVICE_GRPC_URL")
	if statsServiceURL == "" {
		log.Fatal("STATS_SERVICE_GRPC_URL environment variable is not set")
	}

	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(statsServiceURL, connOpts...)
	if err != nil {
		log.Fatalf("Failed to connect to stats service at %s: %v", statsServiceURL, err)
	}
	return statspb.NewStatsServiceClient(conn)
}
