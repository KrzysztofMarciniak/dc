package middleware

import (
	"_discord_bot/db"
	"_discord_bot/models"
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
func IPBlockMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		blocked, err := models.IsIPBlocked(db.DB, ip)
		if err != nil {
			c.Next()
			return
		}

		if blocked {
			c.AbortWithStatusJSON(429, gin.H{
				"error": "Too many failed login attempts. Please try again in an hour.",
			})
			return
		}

		c.Next()
	}
}
