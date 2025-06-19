package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zahartd/social-network/src/services/stats-service/internal/app"
	"github.com/zahartd/social-network/src/services/stats-service/internal/config"
)

func main() {
	cfg := config.Load()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv, err := app.Build(cfg, lis)
	if err != nil {
		log.Fatalf("build server: %v", err)
	}
	reflection.Register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("serve: %v", err)
		}
	}()
	log.Printf("stats-service started on :%s", cfg.GRPCPort)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	srv.GracefulStop()
	log.Println("stats-service stopped")
}
