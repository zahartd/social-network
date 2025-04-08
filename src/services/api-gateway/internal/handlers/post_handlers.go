package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/api-gateway/internal/utils"
)

const UserIDMetadataKey = "x-user-id"

type PostHandler struct {
	postClient postpb.PostServiceClient
}

func NewPostHandler(client postpb.PostServiceClient) *PostHandler {
	if client == nil {
		log.Fatal("PostHandler: postClient cannot be nil")
	}
	return &PostHandler{postClient: client}
}

func createAuthContext(c *gin.Context) (context.Context, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return nil, status.Error(codes.Internal, "internal error: user ID missing after auth middleware")
	}
	err := utils.ValidateUserID(userIDValue)
	if err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{UserIDMetadataKey: userIDValue.(string)})
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	return ctx, nil
}

func parsePagination(c *gin.Context) (page, pageSize int, err error) {
	const defaultPage = 1
	const defaultPageSize = 10

	pageStr := c.DefaultQuery("page", strconv.Itoa(defaultPage))
	pageSizeStr := c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize))

	page, err = utils.ValidatePage(pageStr)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err = utils.ValidatePageSize(pageSizeStr)
	if err != nil {
		return 0, 0, err
	}

	return page, pageSize, nil
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var req postpb.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
		return
	}
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description is required"})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	res, err := h.postClient.CreatePost(ctx, &req)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusCreated, res.Post)
}

func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID parameter is required"})
		return
	}
	err := utils.ValidatePostID(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	res, err := h.postClient.GetPost(ctx, &postpb.GetPostRequest{PostId: postID})
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res.Post)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID parameter is required"})
		return
	}
	err := utils.ValidatePostID(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		PostId: postID,
		Tags:   reqBody.Tags,
	}

	if reqBody.Title != nil {
		grpcReq.Title = *reqBody.Title
		if grpcReq.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title cannot be empty if provided"})
			return
		}
	}
	if reqBody.Description != nil {
		grpcReq.Description = *reqBody.Description
		if grpcReq.Description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "description cannot be empty if provided"})
			return
		}
	}
	if reqBody.IsPrivate != nil {
		grpcReq.IsPrivate = *reqBody.IsPrivate
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	res, err := h.postClient.UpdatePost(ctx, grpcReq)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, res.Post)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	postID := c.Param("postID")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID parameter is required"})
		return
	}
	err := utils.ValidatePostID(postID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	_, err = h.postClient.DeletePost(ctx, &postpb.DeletePostRequest{PostId: postID})
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

func (h *PostHandler) GetMyPosts(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	grpcReq := &postpb.ListMyPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	res, err := h.postClient.ListMyPosts(ctx, grpcReq)
	if err != nil {
		MapGrpcError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) GetAllPublicPosts(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	grpcReq := &postpb.ListPublicPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	res, err := h.postClient.ListPublicPosts(ctx, grpcReq)
	if err != nil {
		MapGrpcError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *PostHandler) GetUserPublicPosts(c *gin.Context) {
	targetUserID := c.Param("userID")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID parameter (:userID) is required"})
		return
	}
	err := utils.ValidateUserID(targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	}

	ctx, err := createAuthContext(c)
	if err != nil {
		MapGrpcError(c, err)
		return
	}

	grpcReq := &postpb.ListPublicPostsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		UserId:   &targetUserID,
	}

	res, err := h.postClient.ListPublicPosts(ctx, grpcReq)
	if err != nil {
		MapGrpcError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
