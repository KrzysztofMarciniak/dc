package router

import (
	controllerError "_discord_bot/controllers/error"
	controllerLogin "_discord_bot/controllers/login"
	controllerLogout "_discord_bot/controllers/logout"
	"_discord_bot/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Engine *gin.Engine
}

func New() *gin.Engine {
	r := gin.New()

	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(middleware.AuthContextMiddleware())
	r.Use(middleware.AuthGateMiddleware("/login", "/error", "/logout"))

	registerRoutes(r)

	return r
}

func registerRoutes(r *gin.Engine) {

	r.GET("/login", controllerLogin.ShowLogin)
	r.POST("/login", controllerLogin.Login)
	r.GET("/logout", controllerLogout.Logout)
	r.GET("/error", controllerError.ShowError)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

}
