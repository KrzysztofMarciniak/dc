package controllers

import (
	"net/http"
	"time"

	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func ShowAdminPanel(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	var attempts []models.LoginAttempt
	if err := db.DB.
		Order("created_at DESC").
		Limit(100).
		Find(&attempts).Error; err != nil {

		c.HTML(http.StatusInternalServerError, "admin-panel.tmpl", gin.H{
			"title": "Admin Panel",
			"auth":  auth,
			"error": "Failed to load login attempts",
		})
		return
	}

	serverTime := time.Now().Format(time.RFC1123)

	uptime := time.Since(startTime).String()

	totalRequests := middleware.TotalRequests()
	var user models.User
	err := db.DB.
		Limit(1).
		Select("uuid").
		Take(&user).Error

	dbStatus := "OK"
	if err != nil {
		dbStatus = "DOWN"
	}

	c.HTML(http.StatusOK, "admin-panel.tmpl", gin.H{
		"title":         "Admin Panel",
		"auth":          auth,
		"isLogged":      true,
		"loginAttempts": attempts,

		"serverTime":    serverTime,
		"uptime":        uptime,
		"totalRequests": totalRequests,
		"dbStatus":      dbStatus,
	})
}
