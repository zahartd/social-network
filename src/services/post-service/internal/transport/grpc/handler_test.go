package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/post-service/internal/auth"
	"github.com/zahartd/social-network/src/services/post-service/internal/domain/service"
	"github.com/zahartd/social-network/src/services/post-service/internal/infrastructure/inmemory"
)

const size = 1 << 20

func newGRPCClient(_ *testing.T) (postpb.PostServiceClient, func()) {
	lis := bufconn.Listen(size)
	repo := inmemory.New()
	svc := service.NewPost(repo, nil, nil, nil, nil)

	s := grpc.NewServer(grpc.UnaryInterceptor(auth.AuthInterceptor))
	postpb.RegisterPostServiceServer(s, New(svc))
	go s.Serve(lis)

	dial := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
	cc, _ := grpc.DialContext(context.Background(),
		"buf",
		grpc.WithContextDialer(dial),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	return postpb.NewPostServiceClient(cc), func() { cc.Close(); s.Stop() }
}

func authCtx(uid string) context.Context {
	return metadata.NewOutgoingContext(context.Background(),
		metadata.Pairs(auth.UserIDMetadataKey, uid))
}

func TestPostServiceRPCs(t *testing.T) {
	c, done := newGRPCClient(t)
	defer done()

	author := uuid.NewString()
	other := uuid.NewString()
	var postID, commentID string

	t.Run("CreatePost", func(t *testing.T) {
		t.Run("ok", func(t *testing.T) {
			out, err := c.CreatePost(authCtx(author), &postpb.CreatePostRequest{Title: "first"})
			assert.NoError(t, err)
			assert.Equal(t, "first", out.Post.Title)
			postID = out.Post.Id
		})
		t.Run("invalid", func(t *testing.T) {
			_, err := c.CreatePost(authCtx(author), &postpb.CreatePostRequest{})
			assert.Equal(t, codes.InvalidArgument, status.Code(err))
		})
	})

	t.Run("GetPost", func(t *testing.T) {
		t.Run("found", func(t *testing.T) {
			_, err := c.GetPost(context.Background(), &postpb.GetPostRequest{PostId: postID})
			assert.NoError(t, err)
		})
		t.Run("missing", func(t *testing.T) {
			_, err := c.GetPost(context.Background(), &postpb.GetPostRequest{PostId: uuid.NewString()})
			assert.Equal(t, codes.NotFound, status.Code(err))
		})
	})

	t.Run("UpdatePost", func(t *testing.T) {
		t.Run("forbidden", func(t *testing.T) {
			_, err := c.UpdatePost(authCtx(other), &postpb.UpdatePostRequest{PostId: postID, Title: "hack"})
			assert.Equal(t, codes.PermissionDenied, status.Code(err))
		})
		t.Run("ok", func(t *testing.T) {
			out, err := c.UpdatePost(authCtx(author), &postpb.UpdatePostRequest{PostId: postID, Title: "second"})
			assert.NoError(t, err)
			assert.Equal(t, "second", out.Post.Title)
		})
	})

	t.Run("Feeds", func(t *testing.T) {
		my, _ := c.ListMyPosts(authCtx(author), &postpb.ListMyPostsRequest{Page: 1, PageSize: 10})
		assert.Len(t, my.Posts, 1)

		all, _ := c.ListPublicPosts(context.Background(), &postpb.ListPublicPostsRequest{Page: 1, PageSize: 10})
		assert.Len(t, all.Posts, 1)

		byUser, _ := c.ListPublicPosts(context.Background(), &postpb.ListPublicPostsRequest{UserId: &author, Page: 1, PageSize: 10})
		assert.Len(t, byUser.Posts, 1)
	})

	t.Run("Metrics", func(t *testing.T) {
		_, err := c.ViewPost(context.Background(), &postpb.ViewPostRequest{PostId: postID})
		assert.NoError(t, err)

		_, err = c.LikePost(authCtx(author), &postpb.LikePostRequest{PostId: postID})
		assert.NoError(t, err)

		_, err = c.UnlikePost(authCtx(author), &postpb.UnlikePostRequest{PostId: postID})
		assert.NoError(t, err)
	})

	t.Run("CommentsFlow", func(t *testing.T) {
		cm, _ := c.AddComment(authCtx(author), &postpb.AddCommentRequest{PostId: postID, Text: "nice"})
		commentID = cm.Comment.Id

		list, _ := c.ListComments(context.Background(), &postpb.ListCommentsRequest{PostId: postID, Page: 1, PageSize: 10})
		assert.Len(t, list.Comments, 1)

		_, _ = c.AddReply(authCtx(other), &postpb.AddReplyRequest{
			PostId: postID, ParentCommentId: commentID, Text: "reply",
		})
		reps, _ := c.ListReplies(context.Background(), &postpb.ListRepliesRequest{ParentCommentId: commentID})
		assert.Len(t, reps.Replies, 1)
	})

	t.Run("DeletePost", func(t *testing.T) {
		_, err := c.DeletePost(authCtx(other), &postpb.DeletePostRequest{PostId: postID})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = c.DeletePost(authCtx(author), &postpb.DeletePostRequest{PostId: postID})
		assert.NoError(t, err)
	})
}
