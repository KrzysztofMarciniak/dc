package router

import (
	controllerAdmin "_discord_bot/controllers/admin"
	controllerError "_discord_bot/controllers/error"
	controllerLogin "_discord_bot/controllers/login"
	controllerLogout "_discord_bot/controllers/logout"
	controllerLogs "_discord_bot/controllers/logs"
	controllers "_discord_bot/controllers/manage"
	controllerPassword "_discord_bot/controllers/password"

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
	r.Use(middleware.IPBlockMiddleware())
	registerRoutes(r)

	return r
}

func registerRoutes(r *gin.Engine) {
	login := r.Group("/login")
	{
		login.GET("", controllerLogin.ShowLogin)
		login.POST("", controllerLogin.Login)
	}

	r.GET("/logout", controllerLogout.Logout)
	r.GET("/error", controllerError.ShowError)
	r.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/error?error=Page+not+found")
	})

	admin_panel := r.Group("/admin-panel")
	{
		admin_panel.GET("", controllerAdmin.ShowAdminPanel)

		db := admin_panel.Group("/db")
		{
			db.GET("/download", controllerAdmin.DownloadDatabase)
			db.POST("/upload", controllerAdmin.UploadDatabase)
		}

		logs := admin_panel.Group("/logs")
		{
			logs.GET("", controllerLogs.ViewGlobalLogs)
			logs.GET("/actor/:uuid", controllerLogs.ViewActorLogs)
			logs.GET("/target/:uuid", controllerLogs.ViewTargetLogs)
		}

		password := admin_panel.Group("/change-password")
		{
			password.GET("", controllerPassword.ShowChangePassword)
			password.POST("", controllerPassword.ChangePassword)
		}

		bots := admin_panel.Group("/bots")
		{
			bots.GET("", controllers.ListBots)
			bots.POST("/create", controllers.CreateBot)

			bots.GET("/:uuid", controllers.GetBot)
			bots.POST("/:uuid/update", controllers.UpdateBot)
			bots.POST("/:uuid/delete", controllers.DeleteBot)
		}
	}
}
