package controllers

import (
	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type createOnMessageRequest struct {
	Trigger   string `form:"trigger" binding:"required"`
	Command   string `form:"command" binding:"required"`
	Reply     string `form:"reply" binding:"required"`
	ChannelID string `form:"channel_id"` // Optional: Empty means all channels
	CSRF      string `form:"csrf" binding:"required"`
}

func ListOnMessages(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	uuid := c.Param("uuid")

	list, err := models.ListOnMessage(db.DB, uuid)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots?error=Failed+to+load")
		return
	}

	csrf, _ := models.NewCSRFToken(db.Secret)

	c.HTML(http.StatusOK, "bot_on_message.tmpl", gin.H{
		"botUUID":    uuid,
		"behaviours": list,
		"csrf":       csrf,
		"auth":       auth,
		"isLogged":   ok,
		"success":    c.Query("success"),
		"error":      c.Query("error"),
	})
}

func CreateOnMessage(c *gin.Context) {
	var req createOnMessageRequest
	uuid := c.Param("uuid")

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?error=Missing+fields")
		return
	}

	if ok, _ := models.ValidateCSRFToken(req.CSRF, db.Secret); !ok {
		c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?error=Invalid+csrf")
		return
	}

	// Clean input and handle the optional ChannelID
	req.ChannelID = strings.TrimSpace(req.ChannelID)

	_, err := models.CreateOnMessage(db.DB, uuid, req.Trigger, req.Command, req.Reply, req.ChannelID)
	if err != nil {
		c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?error=Save+failed")
		return
	}

	c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?success=Added")
}

func DeleteOnMessage(c *gin.Context) {
	uuid := c.Param("uuid")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	csrf := c.PostForm("csrf")

	if ok, _ := models.ValidateCSRFToken(csrf, db.Secret); !ok {
		c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?error=Invalid+csrf")
		return
	}

	err := models.DeleteOnMessage(db.DB, uint(id))
	if err != nil {
		c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?error=Delete+failed")
		return
	}

	c.Redirect(303, "/admin-panel/bots/"+uuid+"/on-message?success=Deleted")
}
