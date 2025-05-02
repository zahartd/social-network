package router

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	postpb "github.com/zahartd/social-network/src/gen/go/post"
	"github.com/zahartd/social-network/src/services/api-gateway/internal/auth"
	"github.com/zahartd/social-network/src/services/api-gateway/internal/handlers"
)

func SetupRouter(postClient postpb.PostServiceClient, userServiceURL *url.URL) *gin.Engine {
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	proxyHandlerFunc := handlers.ProxyHandler(userServiceURL)
	router.POST("/user", proxyHandlerFunc)
	router.GET("/user/login", proxyHandlerFunc)
	userProtected := router.Group("/user")
	userProtected.Use(auth.Middleware())
	{
		userProtected.GET("/logout", proxyHandlerFunc)
		userProtected.GET("/:identifier", proxyHandlerFunc)
		userProtected.PUT("/:identifier", proxyHandlerFunc)
		userProtected.DELETE("/:identifier", proxyHandlerFunc)
	}

	postHandlers := handlers.NewPostHandler(postClient)
	postProtected := router.Group("/posts")
	postProtected.Use(auth.Middleware())
	{
		postProtected.POST("", postHandlers.CreatePost)
		postProtected.GET("/:postID", postHandlers.GetPost)
		postProtected.PUT("/:postID", postHandlers.UpdatePost)
		postProtected.DELETE("/:postID", postHandlers.DeletePost)
		postProtected.GET("/list/my", postHandlers.GetMyPosts)
		postProtected.GET("/list/public", postHandlers.GetAllPublicPosts)
		postProtected.GET("/list/public/:userID", postHandlers.GetUserPublicPosts)
		postProtected.POST("/:postID/view", postHandlers.ViewPost)
		postProtected.POST("/:postID/like", postHandlers.LikePost)
		postProtected.DELETE("/:postID/like", postHandlers.UnlikePost)
		postProtected.GET("/:postID/comments", postHandlers.ListComments)
		postProtected.POST("/:postID/comments", postHandlers.AddComment)
		postProtected.POST("/:postID/comments/:commentID/replies", postHandlers.AddReply)
		postProtected.GET("/:postID/comments/:commentID/replies", postHandlers.ListReplies)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	return router
}
