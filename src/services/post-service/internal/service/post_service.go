package service

import (
	"context"
	"errors"
	"time"

	"github.com/lib/pq"
	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/models"
	"github.com/zahartd/social-network/src/services/post-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostService struct {
	repo repository.PostRepository
}

func NewPostService(repo repository.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func ToProtoPost(post *models.Post) *postpb.Post {
	if post == nil {
		return nil
	}
	return &postpb.Post{
		Id:          post.ID,
		UserId:      post.UserID,
		Title:       post.Title,
		Description: post.Description,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		UpdatedAt:   timestamppb.New(post.UpdatedAt),
		IsPrivate:   post.IsPrivate,
		Tags:        post.Tags,
	}
}

func handleRepoError(err error, operation string, postID string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repository.ErrPostNotFound) {
		return status.Errorf(codes.NotFound, "post %s not found", postID)
	}
	if errors.Is(err, repository.ErrForbidden) {
		return status.Errorf(codes.PermissionDenied, "permission denied")
	}
	return status.Errorf(codes.Internal, "failed to %s post %s: %v", operation, postID, err)
}

func (s *PostService) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*models.Post, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	newPost := &models.Post{
		UserID:      userID,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		IsPrivate:   req.GetIsPrivate(),
		Tags:        pq.StringArray(req.GetTags()),
	}

	postID, err := s.repo.CreatePost(ctx, newPost)
	if err != nil {
		return nil, handleRepoError(err, "create", "")
	}
	newPost.ID = postID

	createdPost, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		newPost.CreatedAt = time.Now()
		newPost.UpdatedAt = newPost.CreatedAt
		return newPost, nil
	}

	return createdPost, nil
}

func (s *PostService) GetPost(ctx context.Context, postID string) (*models.Post, error) {
	if postID == "" {
		return nil, status.Error(codes.InvalidArgument, "post ID is required")
	}

	post, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, handleRepoError(err, "get", postID)
	}

	if post.IsPrivate {
		requestingUserID, err := auth.GetUserIDFromContext(ctx)
		if err != nil && !errors.Is(err, status.Errorf(codes.Unauthenticated, "user ID not found in context")) {
			return nil, err
		}
		if requestingUserID == "" || post.UserID != requestingUserID {
			return nil, status.Errorf(codes.PermissionDenied, "you do not have permission to view this post")
		}
	}

	return post, nil
}

func (s *PostService) UpdatePost(ctx context.Context, req *postpb.UpdatePostRequest) (*models.Post, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	postID := req.GetPostId()
	if postID == "" {
		return nil, status.Error(codes.InvalidArgument, "post ID is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title cannot be empty")
	}

	authorID, err := s.repo.GetPostAuthorID(ctx, postID)
	if err != nil {
		return nil, handleRepoError(err, "check author for update", postID)
	}

	if authorID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "you are not authorized to update this post")
	}

	updatedPostData := &models.Post{
		ID:          postID,
		UserID:      userID,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		IsPrivate:   req.GetIsPrivate(),
		Tags:        pq.StringArray(req.GetTags()),
	}

	err = s.repo.UpdatePost(ctx, updatedPostData)
	if err != nil {
		return nil, handleRepoError(err, "update", postID)
	}

	updatedPost, err := s.repo.GetPostByID(ctx, postID)
	if err != nil {
		updatedPostData.UpdatedAt = time.Now()
		return updatedPostData, nil
	}

	return updatedPost, nil
}

func (s *PostService) DeletePost(ctx context.Context, postID string) error {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	if postID == "" {
		return status.Error(codes.InvalidArgument, "post ID is required")
	}

	err = s.repo.DeletePost(ctx, postID, userID)
	if err != nil {
		return handleRepoError(err, "delete", postID)
	}

	return nil
}

func (s *PostService) ListMyPosts(ctx context.Context, req *postpb.ListMyPostsRequest) ([]*postpb.Post, int, error) {
	userID, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page < 1 {
		return nil, 0, status.Error(codes.InvalidArgument, "page should be a positive number")
	}
	if pageSize < 1 {
		return nil, 0, status.Error(codes.InvalidArgument, "page_size should be a positive number")
	}

	posts, totalCount, err := s.repo.GetUserPosts(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to list user posts: %v", err)
	}

	protoPosts := make([]*postpb.Post, 0, len(posts))
	for _, post := range posts {
		protoPosts = append(protoPosts, ToProtoPost(&post))
	}

	return protoPosts, totalCount, nil
}

func (s *PostService) ListPublicPosts(ctx context.Context, req *postpb.ListPublicPostsRequest) ([]*postpb.Post, int, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page < 1 {
		return nil, 0, status.Error(codes.InvalidArgument, "page should be a positive number")
	}
	if pageSize < 1 {
		return nil, 0, status.Error(codes.InvalidArgument, "page_size should be a positive number")
	}

	filterUserID := req.UserId

	posts, totalCount, err := s.repo.GetPublicPosts(ctx, filterUserID, page, pageSize)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to list public posts: %v", err)
	}

	protoPosts := make([]*postpb.Post, 0, len(posts))
	for _, post := range posts {
		protoPosts = append(protoPosts, ToProtoPost(&post))
	}

	return protoPosts, totalCount, nil
}
