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

type createBotRequest struct {
	Name      string `form:"name" binding:"required,min=3,max=100"`
	Token     string `form:"token" binding:"required"`
	PublicKey string `form:"public_key" binding:"required"`
	CSRF      string `form:"csrf" binding:"required"`
}

type updateBotRequest struct {
	Name string `form:"name" binding:"required,min=3,max=100"`
	CSRF string `form:"csrf" binding:"required"`
}

type deleteBotRequest struct {
	CSRF string `form:"csrf" binding:"required"`
}

func ListBots(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	bots, total, err := models.ListDiscordBots(db.DB, page, 10)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "bots.tmpl", gin.H{
			"error": "Failed to fetch bots",
		})
		return
	}

	csrf, _ := models.NewCSRFToken(db.Secret)

	c.HTML(http.StatusOK, "bots.tmpl", gin.H{
		"title":    "Bots",
		"bots":     bots,
		"total":    total,
		"page":     page,
		"csrf":     csrf,
		"auth":     auth,
		"isLogged": ok,
		"success":  c.Query("success"),
		"error":    c.Query("error"),
	})
}
func CreateBot(c *gin.Context) {
	var req createBotRequest
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots?error=Invalid+input")
		return
	}

	ok, err := models.ValidateCSRFToken(req.CSRF, db.Secret)
	if err != nil || !ok {
		c.Redirect(http.StatusFound, "/admin-panel/bots?error=Invalid+csrf")
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	_, err = models.CreateDiscordBot(db.DB, req.Name, req.Token, req.PublicKey)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots?error=Create+failed")
		return
	}

	c.Redirect(http.StatusFound, "/admin-panel/bots?success=Bot+created")
}
func GetBot(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	uuid := c.Param("uuid")

	bot, err := models.GetDiscordBot(db.DB, uuid)
	if err != nil {
		c.HTML(http.StatusNotFound, "bot.tmpl", gin.H{
			"error": "Bot not found",
		})
		return
	}

	csrf, _ := models.NewCSRFToken(db.Secret)

	c.HTML(http.StatusOK, "bot.tmpl", gin.H{
		"title":    "Bot",
		"bot":      bot,
		"csrf":     csrf,
		"auth":     auth,
		"isLogged": ok,
	})
}
func UpdateBot(c *gin.Context) {
	var req updateBotRequest
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	uuid := c.Param("uuid")

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(400, "/admin-panel/bots/"+uuid+"?error=Invalid+input")
		return
	}

	ok, err := models.ValidateCSRFToken(req.CSRF, db.Secret)
	if err != nil || !ok {
		c.Redirect(400, "/admin-panel/bots/"+uuid+"?error=Invalid+csrf")
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	err = models.UpdateDiscordBotName(db.DB, uuid, req.Name)
	if err != nil {
		c.Redirect(500, "/admin-panel/bots/"+uuid+"?error=Update+failed")
		return
	}

	c.Redirect(303, "/admin-panel/bots/"+uuid+"?success=Updated")
}
func DeleteBot(c *gin.Context) {
	var req deleteBotRequest
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	uuid := c.Param("uuid")

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(400, "/admin-panel/bots?error=Invalid+input")
		return
	}

	ok, err := models.ValidateCSRFToken(req.CSRF, db.Secret)
	if err != nil || !ok {
		c.Redirect(400, "/admin-panel/bots?error=Invalid+csrf")
		return
	}

	err = models.DeleteDiscordBot(db.DB, uuid)
	if err != nil {
		c.Redirect(500, "/admin-panel/bots?error=Delete+failed")
		return
	}

	c.Redirect(303, "/admin-panel/bots?success=Deleted")
}
