package controllers

import (
	"fmt"
	"net/http"
	"time"

	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"_discord_bot/utils"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func ShowAdminPanel(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	page := 1
	pageSize := 5

	if p := c.Query("page"); p != "" {
		_, _ = fmt.Sscanf(p, "%d", &page)
	}

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	attempts, total, err := models.GetLoginAttemptsWithUser(db.DB, pageSize, offset)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "admin-panel.tmpl", gin.H{
			"title": "Admin Panel",
			"error": "Failed to load login attempts",
		})
		return
	}
	for i := range attempts {
		attempts[i].UserAgentPretty = utils.ParseUserAgent(attempts[i].UserAgent)
		attempts[i].CreatedAtPretty = attempts[i].CreatedAt.Format("2006-01-02 15:04:05")
	}
	serverTime := time.Now().Format(time.RFC1123)
	uptime := time.Since(startTime).String()
	totalRequests := middleware.TotalRequests()

	dbStatus := "OK"
	if db.DB.Exec("SELECT 1").Error != nil {
		dbStatus = "DOWN"
	}

	totalPages := 0
	if total > 0 {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}

	c.HTML(http.StatusOK, "admin-panel.tmpl", gin.H{
		"title":         "Admin Panel",
		"auth":          auth,
		"isLogged":      true,
		"loginAttempts": attempts,

		"page":       page,
		"totalPages": totalPages,
		"hasPrev":    page > 1,
		"hasNext":    page < totalPages,

		"serverTime":    serverTime,
		"uptime":        uptime,
		"totalRequests": totalRequests,
		"dbStatus":      dbStatus,
	})
}
