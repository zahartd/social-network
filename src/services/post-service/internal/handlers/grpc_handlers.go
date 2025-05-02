package handlers

import (
	"context"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/service"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type PostGRPCHandler struct {
	postpb.UnimplementedPostServiceServer
	postService *service.PostService
}

func NewPostGRPCHandler(postService *service.PostService) *PostGRPCHandler {
	return &PostGRPCHandler{
		postService: postService,
	}
}

func (h *PostGRPCHandler) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*postpb.PostResponse, error) {
	createdPost, err := h.postService.CreatePost(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(createdPost)}, nil
}

func (h *PostGRPCHandler) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.PostResponse, error) {
	post, err := h.postService.GetPost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(post)}, nil
}

func (h *PostGRPCHandler) UpdatePost(ctx context.Context, req *postpb.UpdatePostRequest) (*postpb.PostResponse, error) {
	updatedPost, err := h.postService.UpdatePost(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.PostResponse{Post: service.ToProtoPost(updatedPost)}, nil
}

func (h *PostGRPCHandler) DeletePost(ctx context.Context, req *postpb.DeletePostRequest) (*emptypb.Empty, error) {
	err := h.postService.DeletePost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *PostGRPCHandler) ListMyPosts(ctx context.Context, req *postpb.ListMyPostsRequest) (*postpb.ListPostsResponse, error) {
	posts, totalCount, err := h.postService.ListMyPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListPostsResponse{
		Posts:      posts,
		TotalCount: int32(totalCount),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *PostGRPCHandler) ListPublicPosts(ctx context.Context, req *postpb.ListPublicPostsRequest) (*postpb.ListPostsResponse, error) {
	posts, totalCount, err := h.postService.ListPublicPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ListPostsResponse{
		Posts:      posts,
		TotalCount: int32(totalCount),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *PostGRPCHandler) ViewPost(ctx context.Context, req *postpb.ViewPostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.postService.ViewPost(ctx, req)
}

func (h *PostGRPCHandler) LikePost(ctx context.Context, req *postpb.LikePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.postService.LikePost(ctx, req)
}

func (h *PostGRPCHandler) UnlikePost(ctx context.Context, req *postpb.UnlikePostRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, h.postService.UnlikePost(ctx, req)
}

func (h *PostGRPCHandler) AddComment(ctx context.Context, req *postpb.AddCommentRequest) (*postpb.CommentResponse, error) {
	cm, err := h.postService.AddComment(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.CommentResponse{Comment: service.ToProtoComment(cm)}, nil
}

func (h *PostGRPCHandler) AddReply(ctx context.Context, req *postpb.AddReplyRequest) (*postpb.ReplyResponse, error) {
	rp, err := h.postService.AddReply(ctx, req)
	if err != nil {
		return nil, err
	}
	return &postpb.ReplyResponse{Reply: service.ToProtoReply(rp)}, nil
}

func (h *PostGRPCHandler) ListComments(ctx context.Context, req *postpb.ListCommentsRequest) (*postpb.ListCommentsResponse, error) {
	cms, total, _ := h.postService.ListComments(ctx, req)
	return &postpb.ListCommentsResponse{Comments: cms, TotalCount: int32(total), Page: req.Page, PageSize: req.PageSize}, nil
}

func (h *PostGRPCHandler) ListReplies(ctx context.Context, req *postpb.ListRepliesRequest) (*postpb.ListRepliesResponse, error) {
	reps, _ := h.postService.ListReplies(ctx, req)
	return &postpb.ListRepliesResponse{Replies: reps, TotalCount: int32(len(reps)), Page: 1, PageSize: int32(len(reps))}, nil
}
