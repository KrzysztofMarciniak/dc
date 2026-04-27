package main

import (
	"net/http"

	"_discord_bot/db"
	"_discord_bot/templates"

	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()

	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(gin.Logger(), gin.Recovery())

	templates.Init(r)
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "template", gin.H{
			"title": "Home",
		})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.Run(":8080")
}
