package handlers

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func ProxyHandler(target *url.URL) gin.HandlerFunc {
	if target == nil {
		log.Fatal("ProxyHandler: target URL cannot be nil")
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
