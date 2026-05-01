package controllers

import (
	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ViewGlobalLogs(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 50

	logs, totalRows, err := models.GetSystemGlobalLogs(db.DB, page, pageSize)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load logs"})
		return
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))

	c.HTML(http.StatusOK, "logs.tmpl", gin.H{
		"User":       auth,
		"Logs":       logs,
		"success":    c.Query("success"),
		"error":      c.Query("error"),
		"auth":       auth,
		"isLogged":   ok,
		"page":       page,
		"totalPages": totalPages,
		"hasPrev":    page > 1,
		"hasNext":    page < totalPages,
	})
}

func ViewActorLogs(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	actorUUID := c.Param("uuid")
	if actorUUID == "" {
		c.Redirect(http.StatusFound, "/admin/logs?error=Actor+UUID+required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 50

	logs, totalRows, err := models.GetLogsByActor(db.DB, actorUUID, page, pageSize)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load actor logs"})
		return
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))

	c.HTML(http.StatusOK, "logs.tmpl", gin.H{
		"User":       auth,
		"Logs":       logs,
		"ActorUUID":  actorUUID,
		"success":    c.Query("success"),
		"error":      c.Query("error"),
		"auth":       auth,
		"isLogged":   ok,
		"page":       page,
		"totalPages": totalPages,
		"hasPrev":    page > 1,
		"hasNext":    page < totalPages,
		"prevPage":   page - 1,
		"nextPage":   page + 1,
	})
}

func ViewTargetLogs(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	targetID := c.Param("uuid") // Matches your route :uuid
	if targetID == "" {
		c.Redirect(http.StatusFound, "/admin/logs?error=Target+ID+required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 50

	logs, totalRows, err := models.GetLogsByTarget(db.DB, targetID, page, pageSize)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load target logs"})
		return
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))

	c.HTML(http.StatusOK, "logs.tmpl", gin.H{
		"User":       auth,
		"Logs":       logs,
		"TargetID":   targetID,
		"success":    c.Query("success"),
		"error":      c.Query("error"),
		"auth":       auth,
		"isLogged":   ok,
		"page":       page,
		"totalPages": totalPages,
		"hasPrev":    page > 1,
		"hasNext":    page < totalPages,
		"prevPage":   page - 1,
		"nextPage":   page + 1,
	})
}
