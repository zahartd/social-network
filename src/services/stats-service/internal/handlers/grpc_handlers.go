package handlers

import (
	"context"

	"github.com/zahartd/social-network/src/services/stats-service/internal/service"

	"github.com/ClickHouse/clickhouse-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
)

type GRPCHandler struct {
	statspb.UnimplementedStatsServiceServer
	svc *service.StatsService
}

func NewGRPCHandler(ckConn clickhouse.Conn) *GRPCHandler {
	return &GRPCHandler{svc: service.NewStatsService(ckConn)}
}

func (h *GRPCHandler) GetPostStats(ctx context.Context, req *statspb.GetPostStatsRequest) (*statspb.GetPostStatsResponse, error) {
	v, l, c, err := h.svc.GetPostStats(req.GetPostId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &statspb.GetPostStatsResponse{Views: int64(v), Likes: int64(l), Comments: int64(c)}, nil
}

func (h *GRPCHandler) GetPostDynamics(ctx context.Context, req *statspb.GetDynamicsRequest) (*statspb.GetDynamicsResponse, error) {
	metric := map[statspb.GetDynamicsRequest_Metric]string{
		statspb.GetDynamicsRequest_VIEWS:    "post-views",
		statspb.GetDynamicsRequest_LIKES:    "post-likes",
		statspb.GetDynamicsRequest_COMMENTS: "post-comments",
	}[req.GetMetric()]
	data, err := h.svc.GetDynamics(req.GetPostId(), metric)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resp := &statspb.GetDynamicsResponse{}
	for _, d := range data {
		resp.Data = append(resp.Data, &statspb.DayCount{Date: d.Date, Count: int64(d.Count)})
	}
	return resp, nil
}

func (h *GRPCHandler) GetTopPosts(ctx context.Context, req *statspb.GetTopRequest) (*statspb.TopPostsResponse, error) {
	by := map[statspb.GetTopRequest_By]string{
		statspb.GetTopRequest_VIEWS:    "post-views",
		statspb.GetTopRequest_LIKES:    "post-likes",
		statspb.GetTopRequest_COMMENTS: "post-comments",
	}[req.GetBy()]
	items, err := h.svc.GetTop(by)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resp := &statspb.TopPostsResponse{}
	for _, it := range items {
		resp.Items = append(resp.Items, &statspb.TopPost{PostId: it.ID, Count: int64(it.Count)})
	}
	return resp, nil
}

func (h *GRPCHandler) GetTopUsers(ctx context.Context, req *statspb.GetTopRequest) (*statspb.TopUsersResponse, error) {
	by := map[statspb.GetTopRequest_By]string{
		statspb.GetTopRequest_VIEWS:    "post-views",
		statspb.GetTopRequest_LIKES:    "post-likes",
		statspb.GetTopRequest_COMMENTS: "post-comments",
	}[req.GetBy()]
	items, err := h.svc.GetTop(by)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	resp := &statspb.TopUsersResponse{}
	for _, it := range items {
		resp.Items = append(resp.Items, &statspb.TopUser{UserId: it.ID, Count: int64(it.Count)})
	}
	return resp, nil
}
