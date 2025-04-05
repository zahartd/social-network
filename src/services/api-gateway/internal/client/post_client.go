package client

import (
	"log"
	"os"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitPostServiceClient() postpb.PostServiceClient {
	postServiceURL := os.Getenv("POST_SERVICE_GRPC_URL")
	if postServiceURL == "" {
		log.Fatal("POST_SERVICE_GRPC_URL environment variable is not set")
	}

	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(postServiceURL, connOpts...)
	if err != nil {
		log.Fatalf("Failed to connect to post service at %s: %v", postServiceURL, err)
	}
	return postpb.NewPostServiceClient(conn)
}
