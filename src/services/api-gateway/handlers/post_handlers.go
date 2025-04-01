package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	postpb "github.com/zahartd/social-network/src/services/api-gateway/pkg/grpc/post"
)

const UserIDMetadataKey = "x-user-id"

type PostHandler struct {
	postClient postpb.PostServiceClient
}

func NewPostHandler(client postpb.PostServiceClient) *PostHandler {
	return &PostHandler{postClient: client}
}

func createAuthContext(c *gin.Context) (context.Context, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		log.Println("Gateway Error: userID not found in Gin context (auth middleware failed?)")
		return nil, status.Error(codes.Internal, "user ID missing after auth")
	}
	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		log.Printf("Gateway Error: Invalid userID type or empty in Gin context: %v (%T)", userIDValue, userIDValue)
		return nil, status.Error(codes.Internal, "invalid user ID in context")
	}

	ctx, _ := context.WithTimeout(c.Request.Context(), 5*time.Second)

	md := metadata.New(map[string]string{UserIDMetadataKey: userID})
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func mapGrpcError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		log.Printf("Gateway Error: Non-gRPC error received: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("Gateway Info: Received gRPC error: code=%s, message=%s", st.Code(), st.Message())

	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	case codes.AlreadyExists:
		c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
	case codes.DeadlineExceeded:
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request to post service timed out"})
	case codes.Unavailable:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Post service is currently unavailable"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Post service internal error"})
	}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req postpb.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authenticated context"})
		return
	}

	log.Printf("Gateway -> gRPC CreatePost: Title='%s'", req.Title)
	res, err := h.postClient.CreatePost(ctx, &req)
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res.Post)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID is required"})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authenticated context"})
		return
	}

	log.Printf("Gateway -> gRPC GetPost: PostID=%s", postID)
	res, err := h.postClient.GetPost(ctx, &postpb.GetPostRequest{PostId: postID})
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res.Post)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID is required"})
		return
	}

	var reqBody struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		IsPrivate   *bool    `json:"is_private"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	grpcReq := &postpb.UpdatePostRequest{
		PostId:      postID,
		Title:       derefString(reqBody.Title, ""),
		Description: derefString(reqBody.Description, ""),
		IsPrivate:   derefBool(reqBody.IsPrivate, false),
		Tags:        reqBody.Tags,
	}
	if grpcReq.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authenticated context"})
		return
	}

	log.Printf("Gateway -> gRPC UpdatePost: PostID=%s", postID)
	res, err := h.postClient.UpdatePost(ctx, grpcReq)
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res.Post)
}

func derefString(s *string, def string) string {
	if s != nil {
		return *s
	}
	return def
}
func derefBool(b *bool, def bool) bool {
	if b != nil {
		return *b
	}
	return def
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID is required"})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authenticated context"})
		return
	}

	log.Printf("Gateway -> gRPC DeletePost: PostID=%s", postID)
	_, err = h.postClient.DeletePost(ctx, &postpb.DeletePostRequest{PostId: postID})
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func (h *PostHandler) ListMyPosts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, errPage := strconv.Atoi(pageStr)
	pageSize, errPageSize := strconv.Atoi(pageSizeStr)

	if errPage != nil || page < 1 {
		page = 1
	}
	if errPageSize != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authenticated context"})
		return
	}

	grpcReq := &postpb.ListUserPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	log.Printf("Gateway -> gRPC ListUserPosts: Page=%d, PageSize=%d", page, pageSize)
	res, err := h.postClient.ListUserPosts(ctx, grpcReq)
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) ListPublicPosts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, errPage := strconv.Atoi(pageStr)
	pageSize, errPageSize := strconv.Atoi(pageSizeStr)

	if errPage != nil || page < 1 {
		page = 1
	}
	if errPageSize != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	grpcReq := &postpb.ListPublicPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	log.Printf("Gateway -> gRPC ListPublicPosts: Page=%d, PageSize=%d", page, pageSize)
	res, err := h.postClient.ListPublicPosts(ctx, grpcReq)
	if err != nil {
		mapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
