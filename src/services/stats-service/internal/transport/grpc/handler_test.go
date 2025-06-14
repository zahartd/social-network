package grpc_handler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/stats-service/internal/infrastructure/inmemory"
)

const bufSize = 1 << 20

func newClient(_ *testing.T) (statspb.StatsServiceClient, func(ev models.Event)) {
	r := inmemory.New()
	svc := service.New(r)
	h := New(svc)

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	statspb.RegisterStatsServiceServer(s, h)
	go s.Serve(lis)

	ctxDialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	cc, _ := grpc.DialContext(context.Background(), "buf",
		grpc.WithContextDialer(ctxDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := statspb.NewStatsServiceClient(cc)
	return client, func(ev models.Event) { _ = r.Insert(context.Background(), ev) }
}

func TestStatsGRPC(t *testing.T) {
	c, push := newClient(t)
	uid := uuid.NewString()
	post := uuid.NewString()
	now := time.Now().UTC()

	push(models.Event{Metric: "post-views", Time: now, UserID: uid, PostID: post, Count: 3})
	push(models.Event{Metric: "post-likes", Time: now, UserID: uid, PostID: post, Count: 2})
	push(models.Event{Metric: "post-unlikes", Time: now, UserID: uid, PostID: post, Count: 1})
	push(models.Event{Metric: "post-comments", Time: now, UserID: uid, PostID: post, Count: 5})

	t.Run("GetPostStats", func(t *testing.T) {
		resp, err := c.GetPostStats(context.Background(), &statspb.GetPostStatsRequest{PostId: post})
		assert.NoError(t, err)
		assert.Equal(t, int64(3), resp.Views)
		assert.Equal(t, int64(1), resp.Likes)
		assert.Equal(t, int64(5), resp.Comments)
	})

	t.Run("Dynamics", func(t *testing.T) {
		resp, err := c.GetPostDynamics(context.Background(), &statspb.GetDynamicsRequest{
			PostId: post, Metric: statspb.GetDynamicsRequest_VIEWS,
		})
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, int64(3), resp.Data[0].Count)
	})

	t.Run("TopPosts", func(t *testing.T) {
		resp, err := c.GetTopPosts(context.Background(), &statspb.GetTopRequest{By: statspb.GetTopRequest_COMMENTS})
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Items)
		assert.Equal(t, post, resp.Items[0].PostId)
	})

	t.Run("TopUsers", func(t *testing.T) {
		resp, err := c.GetTopUsers(context.Background(),
			&statspb.GetTopRequest{By: statspb.GetTopRequest_VIEWS})
		assert.NoError(t, err)

		var got int64
		for _, it := range resp.Items {
			if it.UserId == uid {
				got = it.Count
				break
			}
		}
		assert.Equal(t, int64(3), got)
	})
}
