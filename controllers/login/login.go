package controllers

import (
	"_discord_bot/db"
	"_discord_bot/middleware"
	"_discord_bot/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `form:"username" binding:"required,min=3,max=32"`
	Password string `form:"password" binding:"required,max=72"`
	CSRF     string `form:"csrf" binding:"required"`
}

func ShowLogin(c *gin.Context) {
	auth, ok := middleware.GetAuth(c)
	csrf, err := models.NewCSRFToken(db.Secret)

	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.tmpl", gin.H{
			"title":    "Login",
			"error":    "Internal CSRF generation error",
			"auth":     auth,
			"isLogged": ok,
		})
		return
	}
	c.HTML(http.StatusOK, "login.tmpl", gin.H{
		"title":    "Login",
		"csrf":     csrf,
		"success":  c.Query("success"),
		"error":    c.Query("error"),
		"auth":     auth,
		"isLogged": ok,
	})
}

func Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBind(&req); err != nil {
		c.Redirect(http.StatusFound, "/login?error=Invalid+input")
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	ok, err := models.ValidateCSRFToken(req.CSRF, db.Secret)
	if err != nil || !ok {
		c.Redirect(http.StatusFound, "/login?error=Invalid+csrf+token")
		return
	}

	user, err := models.Authenticate(
		db.DB,
		req.Username,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=Invalid+username+or+password")
		return
	}

	val := user.UUID

	sig, err := models.Sign(val, db.Secret)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=Cookie+Internal+Error")
		return
	}

	cookieVal := val + ":" + sig

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "uid",
		Value:    cookieVal,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	c.Redirect(http.StatusFound, "/admin-panel?success=Logged+in")
}
