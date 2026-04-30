package middleware

import (
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

var totalRequests int64

func TotalRequests() int64 {
	return atomic.LoadInt64(&totalRequests)
}

func TotalRequestsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		atomic.AddInt64(&totalRequests, 1)
		c.Next()
	}
}
