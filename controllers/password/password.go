package controllers

import (
	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type changePasswordRequest struct {
	CurrentPassword string `form:"current_password" binding:"required,max=72"`
	NewPassword     string `form:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `form:"confirm_password" binding:"required,max=72"`
	CSRF            string `form:"csrf" binding:"required"`
}

func ShowChangePassword(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	csrf, err := models.NewCSRFToken(db.Secret)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "change-password.tmpl", gin.H{
			"title": "Change Password",
			"error": "CSRF generation failed",
		})
		return
	}

	c.HTML(http.StatusOK, "change-password.tmpl", gin.H{
		"title":    "Change Password",
		"csrf":     csrf,
		"auth":     auth,
		"isLogged": true,

		"success": c.Query("success"),
		"error":   c.Query("error"),
		"message": c.Query("message"),
	})
}

func ChangePassword(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	if !ok || auth == nil {
		c.Redirect(http.StatusFound, "/login?error=Unauthorized")
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/change-password?error=Invalid+input")
		return
	}

	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	valid, err := models.ValidateCSRFToken(req.CSRF, db.Secret)
	if err != nil || !valid {
		c.Redirect(http.StatusFound, "/change-password?error=Invalid+csrf+token")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.Redirect(http.StatusFound, "/change-password?error=Passwords+do+not+match")
		return
	}

	user, err := models.GetUserByUUID(db.DB, auth.UUID)
	if err != nil {
		c.Redirect(http.StatusFound, "/change-password?error=User+not+found")
		return
	}

	if !user.CheckPassword(req.CurrentPassword) {
		c.Redirect(http.StatusFound, "/change-password?error=Invalid+current+password")
		return
	}

	user.Password = req.NewPassword
	if err := user.HashPassword(); err != nil {
		c.Redirect(http.StatusFound, "/change-password?error=Hashing+failed")
		return
	}

	user.LastPasswordChange = time.Now()

	err = db.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Save(user).Error; err != nil {
			return err
		}
		logEntry := models.InternalLog{
			ActorUUID: auth.UUID,
			ActorType: "user",
			Action:    "CHANGE_PASSWORD",
			TargetID:  user.UUID,
			OldValue:  "REDACTED",
			NewValue:  "REDACTED",
			IPAddress: c.ClientIP(),
		}

		return models.RecordInternalLog(tx, logEntry)
	})

	if err := db.DB.Save(user).Error; err != nil {
		c.Redirect(http.StatusFound, "/change-password?error=Update+failed")
		return
	}

	c.Redirect(http.StatusFound, "/admin-panel?success=Password+updated")
}
