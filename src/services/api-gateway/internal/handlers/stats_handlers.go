package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	statspb "github.com/zahartd/social-network/src/grpc/go/stats"
)

type StatsHandler struct {
	client statspb.StatsServiceClient
}

func NewStatsHandler(client statspb.StatsServiceClient) *StatsHandler {
	return &StatsHandler{client: client}
}

func (h *StatsHandler) GetPostStats(c *gin.Context) {
	postID := c.Param("postID")
	req := &statspb.GetPostStatsRequest{PostId: postID}
	res, err := h.client.GetPostStats(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"views":    res.GetViews(),
		"likes":    res.GetLikes(),
		"comments": res.GetComments(),
	})
}

func (h *StatsHandler) GetPostDynamics(c *gin.Context) {
	postID := c.Param("postID")
	metricStr := c.Query("metric")
	var metric statspb.GetDynamicsRequest_Metric
	switch metricStr {
	case "likes":
		metric = statspb.GetDynamicsRequest_LIKES
	case "comments":
		metric = statspb.GetDynamicsRequest_COMMENTS
	default:
		metric = statspb.GetDynamicsRequest_VIEWS
	}

	req := &statspb.GetDynamicsRequest{
		PostId: postID,
		Metric: metric,
	}
	res, err := h.client.GetPostDynamics(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *StatsHandler) GetTopPosts(c *gin.Context) {
	byStr := c.Query("by")
	var by statspb.GetTopRequest_By
	switch byStr {
	case "likes":
		by = statspb.GetTopRequest_LIKES
	case "comments":
		by = statspb.GetTopRequest_COMMENTS
	default:
		by = statspb.GetTopRequest_VIEWS
	}

	req := &statspb.GetTopRequest{By: by}
	res, err := h.client.GetTopPosts(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *StatsHandler) GetTopUsers(c *gin.Context) {
	byStr := c.Query("by")
	var by statspb.GetTopRequest_By
	switch byStr {
	case "likes":
		by = statspb.GetTopRequest_LIKES
	case "comments":
		by = statspb.GetTopRequest_COMMENTS
	default:
		by = statspb.GetTopRequest_VIEWS
	}

	req := &statspb.GetTopRequest{By: by}
	res, err := h.client.GetTopUsers(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
