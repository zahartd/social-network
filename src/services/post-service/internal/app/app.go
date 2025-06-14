package app

import (
	"github.com/jmoiron/sqlx"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/repository"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/post-service/internal/infrastructure/postgres"
	grpchdl "github.com/zahartd/social-network/src/services/post-service/internal/transport/grpc"
)

type KafkaHub interface {
	Liker() *kafka.Writer
	Unliker() *kafka.Writer
	Viewer() *kafka.Writer
	Commenter() *kafka.Writer
}

func BuildServer(repo repository.Post, hub KafkaHub) *grpc.Server {
	ps := service.NewPost(repo, hub.Viewer(), hub.Liker(), hub.Unliker(), hub.Commenter())
	h := grpchdl.New(ps)

	s := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor))
	postpb.RegisterPostServiceServer(s, h)
	return s
}

func BuildPostgresRepo(db *sqlx.DB) repository.Post { return postgres.New(db) }
