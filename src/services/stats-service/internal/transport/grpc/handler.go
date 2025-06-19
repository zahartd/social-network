package grpc_handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
	"github.com/zahartd/social-network/src/services/stats-service/internal/domain/service"
)

type Handler struct {
	statspb.UnimplementedStatsServiceServer
	svc *service.Service
}

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) GetPostStats(ctx context.Context, r *statspb.GetPostStatsRequest) (*statspb.GetPostStatsResponse, error) {
	v, l, c, err := h.svc.GetPostStats(ctx, r.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &statspb.GetPostStatsResponse{Views: v, Likes: l, Comments: c}, nil
}

func (h *Handler) GetPostDynamics(ctx context.Context, r *statspb.GetDynamicsRequest) (*statspb.GetDynamicsResponse, error) {
	m := metricFromDynamics(r.Metric)
	dd, err := h.svc.GetDynamics(ctx, r.PostId, m)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &statspb.GetDynamicsResponse{}
	for _, d := range dd {
		resp.Data = append(resp.Data, &statspb.DayCount{Date: d.Date, Count: int64(d.Count)})
	}
	return resp, nil
}

func (h *Handler) GetTopPosts(ctx context.Context, r *statspb.GetTopRequest) (*statspb.TopPostsResponse, error) {
	m := metricFromTop(r.By)
	top, err := h.svc.GetTopPosts(ctx, m)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var resp statspb.TopPostsResponse
	for _, it := range top {
		resp.Items = append(resp.Items, &statspb.TopPost{PostId: it.ID, Count: it.Count})
	}
	return &resp, nil
}

func (h *Handler) GetTopUsers(ctx context.Context, r *statspb.GetTopRequest) (*statspb.TopUsersResponse, error) {
	m := metricFromTop(r.By)
	top, err := h.svc.GetTopUsers(ctx, m)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var resp statspb.TopUsersResponse
	for _, it := range top {
		resp.Items = append(resp.Items, &statspb.TopUser{UserId: it.ID, Count: it.Count})
	}
	return &resp, nil
}

func metricFromTop(t statspb.GetTopRequest_By) string {
	switch t {
	case statspb.GetTopRequest_LIKES:
		return "post-likes"
	case statspb.GetTopRequest_COMMENTS:
		return "post-comments"
	default:
		return "post-views"
	}
}

func metricFromDynamics(m statspb.GetDynamicsRequest_Metric) string {
	switch m {
	case statspb.GetDynamicsRequest_LIKES:
		return "post-likes"
	case statspb.GetDynamicsRequest_COMMENTS:
		return "post-comments"
	default:
		return "post-views"
	}
}
