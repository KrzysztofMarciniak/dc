package controllers

import (
	"_discord_bot/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowError(c *gin.Context) {
	error := c.Query("error")

	auth, ok := middleware.GetAuth(c)
	c.HTML(http.StatusOK, "error.tmpl", gin.H{
		"title":    "Error",
		"error":    error,
		"auth":     auth,
		"isLogged": ok,
	})
}
