package grpc

import (
	"context"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	postpb.UnimplementedPostServiceServer
	svc *service.Post
}

func New(svc *service.Post) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.PostResponse, error) {
	p, err := h.svc.CreatePost(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(p)}, nil
}

func (h *Handler) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.PostResponse, error) {
	p, err := h.svc.GetPost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(p)}, nil
}

func (h *Handler) UpdatePost(ctx context.Context, req *postpb.UpdatePostRequest) (*postpb.PostResponse, error) {
	p, err := h.svc.UpdatePost(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(p)}, nil
}

func (h *Handler) DeletePost(ctx context.Context, req *postpb.DeletePostRequest) (*emptypb.Empty, error) {
	if err := h.svc.DeletePost(ctx, req.GetPostId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *Handler) ListMyPosts(ctx context.Context, req *postpb.ListMyPostsRequest) (*postpb.ListPostsResponse, error) {
	ps, total, err := h.svc.ListMyPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListPostsResponse{
		Posts:      ps,
		TotalCount: int32(total),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *Handler) ListPublicPosts(ctx context.Context, req *postpb.ListPublicPostsRequest) (*postpb.ListPostsResponse, error) {
	ps, total, err := h.svc.ListPublicPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListPostsResponse{
		Posts:      ps,
		TotalCount: int32(total),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *Handler) ViewPost(ctx context.Context, req *postpb.ViewPostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.svc.ViewPost(ctx, req)
}

func (h *Handler) LikePost(ctx context.Context, req *postpb.LikePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.svc.LikePost(ctx, req)
}

func (h *Handler) UnlikePost(ctx context.Context, req *postpb.UnlikePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.svc.UnlikePost(ctx, req)
}

func (h *Handler) AddComment(ctx context.Context, req *postpb.AddCommentRequest) (*postpb.CommentResponse, error) {
	cm, err := h.svc.AddComment(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.CommentResponse{Comment: service.ToProtoComment(cm)}, nil
}

func (h *Handler) AddReply(ctx context.Context, req *postpb.AddReplyRequest) (*postpb.ReplyResponse, error) {
	rp, err := h.svc.AddReply(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ReplyResponse{Reply: service.ToProtoReply(rp)}, nil
}

func (h *Handler) ListComments(ctx context.Context, req *postpb.ListCommentsRequest) (*postpb.ListCommentsResponse, error) {
	cms, total, err := h.svc.ListComments(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListCommentsResponse{
		Comments:   cms,
		TotalCount: int32(total),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *Handler) ListReplies(ctx context.Context, req *postpb.ListRepliesRequest) (*postpb.ListRepliesResponse, error) {
	rps, err := h.svc.ListReplies(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListRepliesResponse{
		Replies:    rps,
		TotalCount: int32(len(rps)),
		Page:       1,
		PageSize:   int32(len(rps)),
	}, nil
}
