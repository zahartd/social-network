package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zahartd/social-network/src/services/post-service/internal/app"
	"github.com/zahartd/social-network/src/services/post-service/internal/config"
)

type kafkaHub struct {
	view, like, unlike, comment *kafka.Writer
}

func (h kafkaHub) Viewer() *kafka.Writer    { return h.view }
func (h kafkaHub) Liker() *kafka.Writer     { return h.like }
func (h kafkaHub) Unliker() *kafka.Writer   { return h.unlike }
func (h kafkaHub) Commenter() *kafka.Writer { return h.comment }

func main() {
	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DB_DSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	repo := app.BuildPostgresRepo(db)

	newWriter := func(topic string) *kafka.Writer {
		return &kafka.Writer{
			Addr:                   kafka.TCP(cfg.KafkaBrokerURL),
			Topic:                  topic,
			Async:                  true,
			AllowAutoTopicCreation: true,
		}
	}

	hub := kafkaHub{
		view:    newWriter("post-views"),
		like:    newWriter("post-likes"),
		unlike:  newWriter("post-unlikes"),
		comment: newWriter("post-comments"),
	}

	grpcServer := app.BuildServer(repo, hub)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen %s: %v", cfg.GRPCPort, err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Fatalf("serve: %v", err)
		}
	}()
	log.Printf("post-service gRPC started on :%s", cfg.GRPCPort)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownTimer := time.AfterFunc(10*time.Second, func() {
		log.Println("forced exit")
		os.Exit(0)
	})
	grpcServer.GracefulStop()
	shutdownTimer.Stop()

	_ = hub.view.Close()
	_ = hub.like.Close()
	_ = hub.unlike.Close()
	_ = hub.comment.Close()

	log.Println("post-service stopped gracefully")
}
