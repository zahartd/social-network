package handlers

import (
	"context"
	"log"

	"github.com/zahartd/social-network/src/services/post-service/internal/service"
	postpb "github.com/zahartd/social-network/src/gen/go/post"
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
	log.Printf("gRPC Handler: Received CreatePost request: Title='%s'", req.Title)
	createdPost, err := h.postService.CreatePost(ctx, req)
	if err != nil {
		log.Printf("gRPC Handler Error: CreatePost failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Post created successfully: ID=%s", createdPost.ID)
	return &postpb.PostResponse{Post: service.ToProtoPost(createdPost)}, nil
}

func (h *PostGRPCHandler) GetPost(ctx context.Context, req *postpb.GetPostRequest) (*postpb.PostResponse, error) {
	log.Printf("gRPC Handler: Received GetPost request: PostID=%s", req.PostId)
	post, err := h.postService.GetPost(ctx, req.GetPostId())
	if err != nil {
		log.Printf("gRPC Handler Error: GetPost failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Post retrieved successfully: ID=%s", post.ID)
	return &postpb.PostResponse{Post: service.ToProtoPost(post)}, nil
}

func (h *PostGRPCHandler) UpdatePost(ctx context.Context, req *postpb.UpdatePostRequest) (*postpb.PostResponse, error) {
	log.Printf("gRPC Handler: Received UpdatePost request: PostID=%s", req.PostId)
	updatedPost, err := h.postService.UpdatePost(ctx, req)
	if err != nil {
		log.Printf("gRPC Handler Error: UpdatePost failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Post updated successfully: ID=%s", updatedPost.ID)
	return &postpb.PostResponse{Post: service.ToProtoPost(updatedPost)}, nil
}

func (h *PostGRPCHandler) DeletePost(ctx context.Context, req *postpb.DeletePostRequest) (*emptypb.Empty, error) {
	log.Printf("gRPC Handler: Received DeletePost request: PostID=%s", req.PostId)
	err := h.postService.DeletePost(ctx, req.GetPostId())
	if err != nil {
		log.Printf("gRPC Handler Error: DeletePost failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Post deleted successfully: ID=%s", req.PostId)
	return &emptypb.Empty{}, nil
}

func (h *PostGRPCHandler) ListUserPosts(ctx context.Context, req *postpb.ListUserPostsRequest) (*postpb.ListPostsResponse, error) {
	log.Printf("gRPC Handler: Received ListUserPosts request: Page=%d, PageSize=%d", req.Page, req.PageSize)
	posts, totalCount, err := h.postService.ListUserPosts(ctx, req)
	if err != nil {
		log.Printf("gRPC Handler Error: ListUserPosts failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Found %d posts for user. Total count: %d", len(posts), totalCount)
	return &postpb.ListPostsResponse{
		Posts:      posts,
		TotalCount: int32(totalCount),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}

func (h *PostGRPCHandler) ListPublicPosts(ctx context.Context, req *postpb.ListPublicPostsRequest) (*postpb.ListPostsResponse, error) {
	log.Printf("gRPC Handler: Received ListPublicPosts request: Page=%d, PageSize=%d", req.Page, req.PageSize)
	posts, totalCount, err := h.postService.ListPublicPosts(ctx, req)
	if err != nil {
		log.Printf("gRPC Handler Error: ListPublicPosts failed: %v", err)
		return nil, err
	}
	log.Printf("gRPC Handler: Found %d public posts. Total count: %d", len(posts), totalCount)
	return &postpb.ListPostsResponse{
		Posts:      posts,
		TotalCount: int32(totalCount),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	}, nil
}
