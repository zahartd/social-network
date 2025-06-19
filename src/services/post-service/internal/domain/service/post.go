package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/models"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/repository"
	"github.com/zahartd/social-network/src/services/post-service/internal/utils"
)

type Post struct {
	repo                  repository.Post
	viewW, likeW, unlikeW *kafka.Writer
	commentW              *kafka.Writer
}

func NewPost(r repository.Post, vw, lw, uw, cw *kafka.Writer) *Post {
	return &Post{r, vw, lw, uw, cw}
}

func ToProtoPost(p *models.Post) *postpb.Post {
	if p == nil {
		return nil
	}
	return &postpb.Post{
		Id:          p.ID,
		UserId:      p.UserID,
		Title:       p.Title,
		Description: p.Description,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
		IsPrivate:   p.IsPrivate,
		Tags:        p.Tags,
	}
}

func ToProtoComment(c *models.Comment) *postpb.Comment {
	if c == nil {
		return nil
	}
	return &postpb.Comment{
		Id:        c.ID,
		PostId:    c.PostID,
		UserId:    c.UserID,
		Text:      c.Text,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func ToProtoReply(r *models.Reply) *postpb.Reply {
	if r == nil {
		return nil
	}
	return &postpb.Reply{
		Id:              r.ID,
		PostId:          r.PostID,
		ParentCommentId: r.ParentCommentID,
		UserId:          r.UserID,
		Text:            r.Text,
		CreatedAt:       timestamppb.New(r.CreatedAt),
	}
}

func mapErr(op, id string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrPostNotFound):
		return status.Errorf(codes.NotFound, "post %s not found", id)
	case errors.Is(err, repository.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	default:
		return status.Errorf(codes.Internal, "cannot %s post %s: %v", op, id, err)
	}
}

func (s *Post) CreatePost(ctx context.Context, req *postpb.CreatePostRequest) (*models.Post, error) {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = utils.ValidateUserID(uid); err != nil {
		return nil, err
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title required")
	}
	p := &models.Post{
		UserID:      uid,
		Title:       req.Title,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		Tags:        pq.StringArray(req.Tags),
	}
	id, err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, mapErr("create", "", err)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Post) GetPost(ctx context.Context, id string) (*models.Post, error) {
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "post_id required")
	}
	if err := utils.ValidatePostID(id); err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, mapErr("get", id, err)
	}
	if p.IsPrivate {
		uid, _ := auth.GetUserIDFromContext(ctx)
		if uid == "" || uid != p.UserID {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}
	return p, nil
}

func (s *Post) UpdatePost(ctx context.Context, req *postpb.UpdatePostRequest) (*models.Post, error) {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = utils.ValidateUserID(uid); err != nil {
		return nil, err
	}
	id := req.PostId
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "post_id required")
	}
	if err = utils.ValidatePostID(id); err != nil {
		return nil, err
	}
	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title required")
	}
	author, err := s.repo.AuthorID(ctx, id)
	if err != nil {
		return nil, mapErr("author check", id, err)
	}
	if author != uid {
		return nil, status.Error(codes.PermissionDenied, "not author")
	}
	mp := &models.Post{
		ID:          id,
		UserID:      uid,
		Title:       req.Title,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		Tags:        pq.StringArray(req.Tags),
	}
	if err = s.repo.Update(ctx, mp); err != nil {
		return nil, mapErr("update", id, err)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Post) DeletePost(ctx context.Context, id string) error {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err = utils.ValidateUserID(uid); err != nil {
		return err
	}
	if id == "" {
		return status.Error(codes.InvalidArgument, "post_id required")
	}
	if err = utils.ValidatePostID(id); err != nil {
		return err
	}
	return mapErr("delete", id, s.repo.Delete(ctx, id, uid))
}

func (s *Post) ListMyPosts(ctx context.Context, req *postpb.ListMyPostsRequest) ([]*postpb.Post, int, error) {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	if err = utils.ValidateUserID(uid); err != nil {
		return nil, 0, err
	}
	page, err := utils.ValidatePage(strconv.Itoa(int(req.Page)))
	if err != nil {
		return nil, 0, err
	}
	size, err := utils.ValidatePageSize(strconv.Itoa(int(req.PageSize)))
	if err != nil {
		return nil, 0, err
	}
	list, total, err := s.repo.UserFeed(ctx, uid, page, size)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "user feed: %v", err)
	}
	var out []*postpb.Post
	for _, p := range list {
		out = append(out, ToProtoPost(&p))
	}
	return out, total, nil
}

func (s *Post) ListPublicPosts(ctx context.Context, req *postpb.ListPublicPostsRequest) ([]*postpb.Post, int, error) {
	page, err := utils.ValidatePage(strconv.Itoa(int(req.Page)))
	if err != nil {
		return nil, 0, err
	}
	size, err := utils.ValidatePageSize(strconv.Itoa(int(req.PageSize)))
	if err != nil {
		return nil, 0, err
	}
	var filter *string
	if req.UserId != nil && *req.UserId != "" {
		filter = req.UserId
	}
	list, total, err := s.repo.PublicFeed(ctx, filter, page, size)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "public feed: %v", err)
	}
	var out []*postpb.Post
	for _, p := range list {
		out = append(out, ToProtoPost(&p))
	}
	return out, total, nil
}

func (s *Post) ViewPost(ctx context.Context, req *postpb.ViewPostRequest) error {
	uid, _ := auth.GetUserIDFromContext(ctx)
	_ = s.repo.RecordView(ctx, uid, req.PostId)

	ev := map[string]any{
		"user_id":   uid,
		"post_id":   req.PostId,
		"viewed_at": time.Now().UTC(),
	}
	sendKafka(ctx, s.viewW, uid, ev)
	return nil
}

func (s *Post) LikePost(ctx context.Context, req *postpb.LikePostRequest) error {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err = s.repo.RecordLike(ctx, uid, req.PostId); err != nil {
		return err
	}

	ev := map[string]any{
		"user_id":  uid,
		"post_id":  req.PostId,
		"liked_at": time.Now().UTC(),
	}
	sendKafka(ctx, s.likeW, uid, ev)
	return nil
}

func (s *Post) UnlikePost(ctx context.Context, req *postpb.UnlikePostRequest) error {
	uid, err := auth.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	if err = s.repo.RemoveLike(ctx, uid, req.PostId); err != nil {
		return err
	}

	ev := map[string]any{
		"user_id":    uid,
		"post_id":    req.PostId,
		"unliked_at": time.Now().UTC(),
	}
	sendKafka(ctx, s.unlikeW, uid, ev)
	return nil
}

func (s *Post) AddComment(ctx context.Context, req *postpb.AddCommentRequest) (*models.Comment, error) {
	uid, _ := auth.GetUserIDFromContext(ctx)
	cm := &models.Comment{
		PostID: req.PostId,
		UserID: uid,
		Text:   req.Text,
	}
	id, _ := s.repo.CreateComment(ctx, cm)
	cm.ID = id

	ev := map[string]any{
		"user_id":    uid,
		"post_id":    req.PostId,
		"comment_id": id,
		"created_at": time.Now().UTC(),
	}
	sendKafka(ctx, s.commentW, uid, ev)
	return cm, nil
}

func (s *Post) AddReply(ctx context.Context, req *postpb.AddReplyRequest) (*models.Reply, error) {
	uid, _ := auth.GetUserIDFromContext(ctx)
	rp := &models.Reply{
		PostID:          req.PostId,
		ParentCommentID: req.ParentCommentId,
		UserID:          uid,
		Text:            req.Text,
	}
	id, _ := s.repo.CreateReply(ctx, rp)
	rp.ID = id

	ev := map[string]any{
		"user_id":    uid,
		"post_id":    req.PostId,
		"comment_id": id,
		"created_at": time.Now().UTC(),
	}
	sendKafka(ctx, s.commentW, uid, ev)
	return rp, nil
}

func (s *Post) ListComments(ctx context.Context, req *postpb.ListCommentsRequest) ([]*postpb.Comment, int, error) {
	cms, total, err := s.repo.ListComments(ctx, req.PostId, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, 0, err
	}
	var out []*postpb.Comment
	for _, c := range cms {
		out = append(out, ToProtoComment(&c))
	}
	return out, total, nil
}

func (s *Post) ListReplies(ctx context.Context, req *postpb.ListRepliesRequest) ([]*postpb.Reply, error) {
	reps, err := s.repo.ListReplies(ctx, req.ParentCommentId)
	if err != nil {
		return nil, err
	}
	var out []*postpb.Reply
	for _, r := range reps {
		out = append(out, ToProtoReply(&r))
	}
	return out, nil
}

var retryable = map[error]struct{}{
	kafka.LeaderNotAvailable:      {},
	kafka.UnknownTopicOrPartition: {},
}

func sendKafka(ctx context.Context, w *kafka.Writer, key string, ev any) {
	if w == nil {
		return
	}
	b, _ := json.Marshal(ev)

	for range 5 {
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := w.WriteMessages(cctx, kafka.Message{Key: []byte(key), Value: b})
		cancel()
		switch {
		case err == nil:
			return
		case errors.Is(err, context.DeadlineExceeded):
		default:
			if _, ok := retryable[err]; !ok {
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}
