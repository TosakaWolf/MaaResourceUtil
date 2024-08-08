package router

import (
	"github.com/labstack/echo/v4"
	"maaResourceUtil/server/internal/config"
	"maaResourceUtil/server/pkg/api"
)

func InitRouter(e *echo.Echo) {
	initBotRoutes(e.Group("/maa"))
}
func initBotRoutes(g *echo.Group) {
	if config.Config.Cloud189.Enabled {
		g.GET("/getResource", api.GetResource)
	}
}
