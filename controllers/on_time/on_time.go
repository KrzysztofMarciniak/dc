package controllers

import (
	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type createOnTimeRequest struct {
	ChannelID string `form:"channel_id" binding:"required"`
	Message   string `form:"message" binding:"required"`
	TimeType  string `form:"time_type" binding:"required"` // "interval" or "specific"
	// These come as strings from the form
	IntervalValue int64  `form:"interval_value"`
	SpecificDate  string `form:"specific_date"` // e.g., "2026-05-07T14:30"
	CSRF          string `form:"csrf" binding:"required"`
}

func ListOnTime(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	uuid := c.Param("uuid")

	list, err := models.ListOnTime(db.DB, uuid)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots?error=Failed+to+load")
		return
	}

	csrf, _ := models.NewCSRFToken(db.Secret)

	c.HTML(http.StatusOK, "bot_on_time.tmpl", gin.H{
		"botUUID":    uuid,
		"behaviours": list,
		"csrf":       csrf,
		"auth":       auth,
		"isLogged":   ok,
		"success":    c.Query("success"),
		"error":      c.Query("error"),
	})
}

func CreateOnTime(c *gin.Context) {
	var req createOnTimeRequest
	uuid := c.Param("uuid")

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Missing+fields")
		return
	}

	if ok, _ := models.ValidateCSRFToken(req.CSRF, db.Secret); !ok {
		c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Invalid+csrf")
		return
	}

	var finalTimeValue int64

	if req.TimeType == "interval" {
		// IntervalValue is seconds (e.g., user wants 1 hour = 3600)
		finalTimeValue = req.IntervalValue
	} else {
		// Parse "2026-05-07T14:30" (standard datetime-local format)
		t, err := time.Parse("2006-01-02T15:04", req.SpecificDate)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Invalid+date+format")
			return
		}
		finalTimeValue = t.Unix()
	}

	_, err := models.CreateOnTime(db.DB, uuid, req.ChannelID, req.Message, req.TimeType, finalTimeValue)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Save+failed")
		return
	}

	c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?success=Schedule+added")
}

func DeleteOnTime(c *gin.Context) {
	uuid := c.Param("uuid")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	csrf := c.PostForm("csrf")

	if ok, _ := models.ValidateCSRFToken(csrf, db.Secret); !ok {
		c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Invalid+csrf")
		return
	}

	if err := models.DeleteOnTime(db.DB, uint(id)); err != nil {
		c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?error=Delete+failed")
		return
	}

	c.Redirect(http.StatusFound, "/admin-panel/bots/"+uuid+"/on-time?success=Deleted")
}
