package handlers

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
	"github.com/zahartd/social-network/src/services/stats-service/internal/service"
)

type GRPCHandler struct {
	statspb.UnimplementedStatsServiceServer
	svc *service.StatsService
}

func NewGRPCHandler(ck clickhouse.Conn) *GRPCHandler {
	return &GRPCHandler{svc: service.NewStatsService(ck)}
}

func (h *GRPCHandler) GetPostStats(ctx context.Context, req *statspb.GetPostStatsRequest) (*statspb.GetPostStatsResponse, error) {
	views, likes, comments, err := h.svc.GetPostStats(ctx, req.GetPostId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetPostStats: %v", err)
	}
	return &statspb.GetPostStatsResponse{
		Views:    views,
		Likes:    likes,
		Comments: comments,
	}, nil
}

func (h *GRPCHandler) GetPostDynamics(ctx context.Context, req *statspb.GetDynamicsRequest) (*statspb.GetDynamicsResponse, error) {
	var metric string
	switch req.GetMetric() {
	case statspb.GetDynamicsRequest_LIKES:
		metric = "post-likes"
	case statspb.GetDynamicsRequest_COMMENTS:
		metric = "post-comments"
	default:
		metric = "post-views"
	}

	data, err := h.svc.GetDynamics(ctx, req.GetPostId(), metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetPostDynamics: %v", err)
	}

	resp := &statspb.GetDynamicsResponse{}
	for _, d := range data {
		resp.Data = append(resp.Data, &statspb.DayCount{
			Date:  d.Date,
			Count: int64(d.Count),
		})
	}
	return resp, nil
}

func (h *GRPCHandler) GetTopPosts(ctx context.Context, req *statspb.GetTopRequest) (*statspb.TopPostsResponse, error) {
	var metric string
	switch req.GetBy() {
	case statspb.GetTopRequest_LIKES:
		metric = "post-likes"
	case statspb.GetTopRequest_COMMENTS:
		metric = "post-comments"
	default:
		metric = "post-views"
	}

	items, err := h.svc.GetTopPosts(ctx, metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetTopPosts: %v", err)
	}

	resp := &statspb.TopPostsResponse{}
	for _, it := range items {
		resp.Items = append(resp.Items, &statspb.TopPost{
			PostId: it.ID,
			Count:  it.Count,
		})
	}
	return resp, nil
}

func (h *GRPCHandler) GetTopUsers(ctx context.Context, req *statspb.GetTopRequest) (*statspb.TopUsersResponse, error) {
	var metric string
	switch req.GetBy() {
	case statspb.GetTopRequest_LIKES:
		metric = "post-likes"
	case statspb.GetTopRequest_COMMENTS:
		metric = "post-comments"
	default:
		metric = "post-views"
	}

	items, err := h.svc.GetTopUsers(ctx, metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "GetTopUsers: %v", err)
	}

	resp := &statspb.TopUsersResponse{}
	for _, it := range items {
		resp.Items = append(resp.Items, &statspb.TopUser{
			UserId: it.ID,
			Count:  it.Count,
		})
	}
	return resp, nil
}
