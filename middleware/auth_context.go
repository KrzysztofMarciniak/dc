package middleware

import (
	"strings"

	"_discord_bot/db"
	"_discord_bot/models"

	"github.com/gin-gonic/gin"
)

type AuthContext struct {
	UUID     string
	Username string
}

const AuthKey = "auth"

func AuthContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("uid")
		if err != nil || cookie == "" {
			c.Next()
			return
		}

		parts := strings.Split(cookie, ":")
		if len(parts) != 2 {
			c.Next()
			return
		}

		userUUID := parts[0]
		sig := parts[1]

		ok, err := models.Verify(userUUID, sig, db.Secret)
		if err != nil || !ok {
			c.Next()
			return
		}

		user, err := models.GetUserByUUID(db.DB, userUUID)
		if err == nil {
			c.Set("auth", AuthContext{
				UUID:     user.UUID,
				Username: user.Username,
			})
		}

		c.Next()
	}
}

func GetAuth(c *gin.Context) (*AuthContext, bool) {
	v, exists := c.Get(AuthKey)
	if !exists || v == nil {
		return nil, false
	}

	ac, ok := v.(AuthContext)
	if !ok {
		return nil, false
	}

	return &ac, true
}
func AuthGateMiddleware(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, p := range allowed {
		allowedSet[p] = struct{}{}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		if _, ok := allowedSet[path]; ok {
			c.Next()
			return
		}

		// ONLY use context here
		if _, exists := c.Get(AuthKey); !exists {
			c.Redirect(302, "/error?error=Please+log+in")
			c.Abort()
			return
		}

		c.Next()
	}
}
