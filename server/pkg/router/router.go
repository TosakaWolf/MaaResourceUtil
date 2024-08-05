package router

import (
	"github.com/labstack/echo/v4"
	"maaResFetch/server/pkg/api"
)

func InitRouter(e *echo.Echo) {
	initBotRoutes(e.Group("/maa"))
}
func initBotRoutes(g *echo.Group) {
	g.GET("/resource", api.GetResource)
}
