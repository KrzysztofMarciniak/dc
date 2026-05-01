package router

import (
	controllerAdmin "_discord_bot/controllers/admin"
	controllerError "_discord_bot/controllers/error"
	controllerLogin "_discord_bot/controllers/login"
	controllerLogout "_discord_bot/controllers/logout"
	controllerLogs "_discord_bot/controllers/logs"
	controllerPassword "_discord_bot/controllers/password"

	controllerRoot "_discord_bot/controllers/root"
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

	r.Use(middleware.TotalRequestsMiddleware())
	r.Use(middleware.AuthContextMiddleware())
	r.Use(middleware.AuthGateMiddleware("/login", "/error", "/logout"))
	r.Use(middleware.PasswordPolicyMiddleware())

	registerRoutes(r)

	return r
}

func registerRoutes(r *gin.Engine) {

	r.GET("/login", controllerLogin.ShowLogin)
	r.POST("/login", controllerLogin.Login)
	r.GET("/logout", controllerLogout.Logout)
	r.GET("/error", controllerError.ShowError)
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/error?error=Page+not+found")
	})
	r.GET("/admin-panel", controllerAdmin.ShowAdminPanel)

	r.GET("/admin/db/download", controllerAdmin.DownloadDatabase)
	r.POST("/admin/db/upload", controllerAdmin.UploadDatabase)

	r.GET("/admin/logs", controllerLogs.ViewGlobalLogs)
	r.GET("/admin/logs/actor/:uuid", controllerLogs.ViewActorLogs)
	r.GET("/admin/logs/target/:uuid", controllerLogs.ViewTargetLogs)

	r.GET("/change-password", controllerPassword.ShowChangePassword)
	r.POST("/change-password", controllerPassword.ChangePassword)
	r.GET("/", controllerRoot.Index)

}
