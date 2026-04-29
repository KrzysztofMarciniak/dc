package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowError(c *gin.Context) {
	message := c.Query("message")

	c.HTML(http.StatusOK, "template", gin.H{
		"title":   "Error",
		"message": message,
	})
}
