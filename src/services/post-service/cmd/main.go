package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/config"
	"github.com/zahartd/social-network/src/services/post-service/internal/handlers"
	"github.com/zahartd/social-network/src/services/post-service/internal/repository"
	"github.com/zahartd/social-network/src/services/post-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DB_DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	postRepo := repository.NewPostgresPostRepository(db)

	viewWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokerURL),
		Topic:                  "post-views",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	defer func() {
		if err := viewWriter.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()
	likeWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokerURL),
		Topic:                  "post-likes",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	defer func() {
		if err := likeWriter.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()
	commentWriter := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokerURL),
		Topic:                  "post-comments",
		Async:                  true,
		AllowAutoTopicCreation: true,
	}
	defer func() {
		if err := commentWriter.Close(); err != nil {
			log.Fatal("failed to close writer:", err)
		}
	}()

	postService := service.NewPostService(postRepo, viewWriter, likeWriter, commentWriter)
	postHandler := handlers.NewPostGRPCHandler(postService)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthInterceptor),
	)

	postpb.RegisterPostServiceServer(grpcServer, postHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	grpcServer.GracefulStop()
}
