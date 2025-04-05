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
