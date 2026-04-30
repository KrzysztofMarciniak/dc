package controllers

import (
	"_discord_bot/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	_, ok := middleware.GetAuth(c)

	if ok {
		c.Redirect(http.StatusFound, "/admin-panel")
		return
	}

	c.Redirect(http.StatusFound, "/login")
}
