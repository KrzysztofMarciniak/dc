package main

import (
	"net/http"

	"dc/db"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.Run(":8080")
}
