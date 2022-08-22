package router

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/websocket"
	"golang-demo/internal/handler"
	"golang-demo/internal/middleware"
)

func RouteDemo(app *iris.Application) {
	// 常用的iris 中间件
	app.UseGlobal(middleware.CROS, middleware.LogInfoBefore, middleware.LogInfoAfter)

	// iris 路由管理
	home := new(handler.HomeCtl)
	home.DeepEqual()
	app.PartyFunc("/home", func(p router.Party) {
		p.Get("/1", home.Test)
	})

	app.Get("/test", func(ctx iris.Context) {
		handler.Re(ctx, handler.Success, nil)
	})

	// websoket 使用
	app.Get("/msg", websocket.Handler(handler.InitWebsocket()))

}
