package controllers

import (
	"_discord_bot/models"

	"github.com/gin-gonic/gin"
)

type createBotRequest struct {
	Name      string `form:"name" binding:"required"`
	Token     string `form:"token" binding:"required"`
	PublicKey string `form:"public_key" binding:"required"`
}

type updateBotRequest struct {
	Name string `form:"name" binding:"required"`
}

func CreateBot(c *gin.Context) {
	var req createBotRequest

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(302, "/admin-panel/bots?error=invalid+input")
		return
	}

	_, err := models.CreateDiscordBot(
		db.Get(), // adjust if needed
		req.Name,
		req.Token,
		req.PublicKey,
	)

	if err != nil {
		c.Redirect(302, "/admin-panel/bots?error=create+failed")
		return
	}

	c.Redirect(302, "/admin-panel/bots?success=created")
}
