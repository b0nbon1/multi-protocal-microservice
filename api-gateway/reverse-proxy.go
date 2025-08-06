package main

import (
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

func reverseProxy(target *url.URL) gin.HandlerFunc {
    return func(c *gin.Context) {
        proxy := httputil.NewSingleHostReverseProxy(target)
        
        // Modify the request path
        c.Request.URL.Host = target.Host
        c.Request.URL.Scheme = target.Scheme
        c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
        c.Request.Host = target.Host

        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
