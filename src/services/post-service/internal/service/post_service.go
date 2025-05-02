package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/models"
	"github.com/zahartd/social-network/src/services/post-service/internal/repository"
	"github.com/zahartd/social-network/src/services/post-service/internal/utils"
)

type PostService struct {
	repo          repository.PostRepository
	viewWriter    *kafka.Writer
	likeWriter    *kafka.Writer
	commentWriter *kafka.Writer
}

func NewPostService(r repository.PostRepository, vw, lw, cw *kafka.Writer) *PostService {
	return &PostService{repo: r, viewWriter: vw, likeWriter: lw, commentWriter: cw}
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

func ToProtoComment(cm *models.Comment) *postpb.Comment {
	if cm == nil {
		return nil
	}
	return &postpb.Comment{
		Id:        cm.ID,
		PostId:    cm.PostID,
		UserId:    cm.UserID,
		Text:      cm.Text,
		CreatedAt: timestamppb.New(cm.CreatedAt),
	}
}

func ToProtoReply(rp *models.Reply) *postpb.Reply {
	if rp == nil {
		return nil
	}
	return &postpb.Reply{
		Id:              rp.ID,
		PostId:          rp.PostID,
		ParentCommentId: rp.ParentCommentID,
		UserId:          rp.UserID,
		Text:            rp.Text,
		CreatedAt:       timestamppb.New(rp.CreatedAt),
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
	err = utils.ValidateUserID(userID)
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
	err := utils.ValidatePostID(postID)
	if err != nil {
		return nil, err
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
	err = utils.ValidateUserID(userID)
	if err != nil {
		return nil, err
	}

	postID := req.GetPostId()
	if postID == "" {
		return nil, status.Error(codes.InvalidArgument, "post ID is required")
	}
	err = utils.ValidatePostID(postID)
	if err != nil {
		return nil, err
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
	err = utils.ValidateUserID(userID)
	if err != nil {
		return err
	}

	if postID == "" {
		return status.Error(codes.InvalidArgument, "post ID is required")
	}
	err = utils.ValidatePostID(postID)
	if err != nil {
		return err
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
	err = utils.ValidateUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	page := int(req.GetPage())
	_, err = utils.ValidatePage(strconv.Itoa(page))
	if err != nil {
		return nil, 0, err
	}
	pageSize := int(req.GetPageSize())
	_, err = utils.ValidatePage(strconv.Itoa(pageSize))
	if err != nil {
		return nil, 0, err
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
	_, err := utils.ValidatePage(strconv.Itoa(page))
	if err != nil {
		return nil, 0, err
	}
	pageSize := int(req.GetPageSize())
	_, err = utils.ValidatePage(strconv.Itoa(pageSize))
	if err != nil {
		return nil, 0, err
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

func (s *PostService) ViewPost(ctx context.Context, req *postpb.ViewPostRequest) error {
	userID, _ := auth.GetUserIDFromContext(ctx)
	_ = s.repo.RecordView(ctx, userID, req.PostId)

	ev := struct {
		UserID   string    `json:"user_id"`
		PostId   string    `json:"post_id"`
		ViewedAt time.Time `json:"viewed_at"`
	}{
		UserID:   userID,
		PostId:   req.PostId,
		ViewedAt: time.Now().UTC(),
	}
	payload, _ := json.Marshal(ev)

	const retries = 3
	for range retries {
		writerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := s.viewWriter.WriteMessages(
			writerCtx,
			kafka.Message{
				Key:   []byte(userID),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Printf("failed to write messages: %s", err.Error())
		}
		break
	}
	return nil
}

func (s *PostService) LikePost(ctx context.Context, req *postpb.LikePostRequest) error {
	userID, _ := auth.GetUserIDFromContext(ctx)
	_ = s.repo.RecordLike(ctx, userID, req.PostId)

	ev := struct {
		UserID  string    `json:"user_id"`
		PostId  string    `json:"post_id"`
		LikedAt time.Time `json:"liked_at"`
	}{
		UserID:  userID,
		PostId:  req.PostId,
		LikedAt: time.Now().UTC(),
	}
	payload, _ := json.Marshal(ev)

	const retries = 3
	for range retries {
		writerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := s.likeWriter.WriteMessages(
			writerCtx,
			kafka.Message{
				Key:   []byte(userID),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Printf("failed to write messages: %s", err.Error())
		}
		break
	}
	return nil
}

func (s *PostService) UnlikePost(ctx context.Context, req *postpb.UnlikePostRequest) error {
	userID, _ := auth.GetUserIDFromContext(ctx)
	_ = s.repo.RemoveLike(ctx, userID, req.PostId)
	return nil
}

func (s *PostService) AddComment(ctx context.Context, req *postpb.AddCommentRequest) (*models.Comment, error) {
	userID, _ := auth.GetUserIDFromContext(ctx)
	cm := &models.Comment{
		PostID: req.PostId,
		UserID: userID,
		Text:   req.Text,
	}
	id, _ := s.repo.CreateComment(ctx, cm)
	cm.ID = id

	ev := struct {
		UserID    string    `json:"user_id"`
		PostId    string    `json:"post_id"`
		CommentId string    `json:"comment_id"`
		CreatedAt time.Time `json:"created_at"`
	}{
		UserID:    userID,
		PostId:    req.PostId,
		CommentId: id,
		CreatedAt: time.Now().UTC(),
	}
	payload, _ := json.Marshal(ev)

	const retries = 3
	for range retries {
		writerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := s.commentWriter.WriteMessages(
			writerCtx,
			kafka.Message{
				Key:   []byte(userID),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Printf("failed to write messages: %s", err.Error())
		}
		break
	}
	return cm, nil
}

func (s *PostService) AddReply(ctx context.Context, req *postpb.AddReplyRequest) (*models.Reply, error) {
	userID, _ := auth.GetUserIDFromContext(ctx)
	rp := &models.Reply{
		PostID:          req.PostId,
		ParentCommentID: req.ParentCommentId,
		UserID:          userID,
		Text:            req.Text,
	}
	id, _ := s.repo.CreateReply(ctx, rp)
	rp.ID = id

	ev := struct {
		UserID    string    `json:"user_id"`
		PostId    string    `json:"post_id"`
		CommentId string    `json:"comment_id"`
		CreatedAt time.Time `json:"created_at"`
	}{
		UserID:    userID,
		PostId:    req.PostId,
		CommentId: id,
		CreatedAt: time.Now().UTC(),
	}
	payload, _ := json.Marshal(ev)

	const retries = 3
	for range retries {
		writerCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := s.commentWriter.WriteMessages(
			writerCtx,
			kafka.Message{
				Key:   []byte(userID),
				Value: payload,
			},
		)
		if errors.Is(err, kafka.LeaderNotAvailable) || errors.Is(err, context.DeadlineExceeded) {
			time.Sleep(time.Millisecond * 250)
			continue
		}

		if err != nil {
			log.Printf("failed to write messages: %s", err.Error())
		}
		break
	}
	return rp, nil
}

func (s *PostService) ListComments(ctx context.Context, req *postpb.ListCommentsRequest) ([]*postpb.Comment, int, error) {
	comments, total, _ := s.repo.ListComments(ctx, req.PostId, int(req.Page), int(req.PageSize))
	var r []*postpb.Comment
	for _, cm := range comments {
		r = append(r, &postpb.Comment{
			Id: cm.ID, PostId: cm.PostID, UserId: cm.UserID,
			Text: cm.Text, CreatedAt: timestamppb.New(cm.CreatedAt),
		})
	}
	return r, total, nil
}

func (s *PostService) ListReplies(ctx context.Context, req *postpb.ListRepliesRequest) ([]*postpb.Reply, error) {
	reps, _ := s.repo.ListReplies(ctx, req.ParentCommentId)
	var r []*postpb.Reply
	for _, rp := range reps {
		r = append(r, &postpb.Reply{
			Id: rp.ID, PostId: rp.PostID, ParentCommentId: rp.ParentCommentID, UserId: rp.UserID,
			Text: rp.Text, CreatedAt: timestamppb.New(rp.CreatedAt),
		})
	}
	return r, nil
}
