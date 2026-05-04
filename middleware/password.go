package middleware

import (
	"_discord_bot/db"
	"_discord_bot/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func PasswordPolicyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		v, exists := c.Get(AuthKey)
		if !exists || v == nil {
			c.Next()
			return
		}

		auth, ok := v.(AuthContext)
		if !ok {
			c.Next()
			return
		}

		user, err := models.GetUserByUUID(db.DB, auth.UUID)
		if err != nil {
			c.Next()
			return
		}
		if time.Since(user.LastPasswordChange) > models.PasswordTimeLeft(user) {

			path := c.Request.URL.Path

			if path != "/admin-panel/change-password" && path != "/logout" {

				c.Redirect(http.StatusFound, "/admin-panel/change-password?error=Password+expired")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
