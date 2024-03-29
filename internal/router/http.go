package router

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/websocket"
	"golang-demo/internal/consts"
	"golang-demo/internal/handler"
	"golang-demo/internal/logic"
	"golang-demo/internal/middleware"
)

func RouteDemo(app *iris.Application) {
	// 常用的iris 中间件
	app.UseGlobal(middleware.CROS, middleware.LogInfoBefore, middleware.LogInfoAfter)

	// iris 路由管理
	home := new(handler.HomeCtl)
	app.PartyFunc("/home", func(p router.Party) {
		p.Get("/1", home.Test)
		p.Get("/query", home.Query)
	})

	app.PartyFunc("/test", func(p router.Party) {
		p.Get("/1", func(ctx iris.Context) {
			handler.Re(ctx, consts.Success, nil)
		})
		p.Get("/2", func(ctx iris.Context) {
			handler.Re(ctx, consts.Success, nil)
		})

		p.Get("/3", func(ctx iris.Context) {
			handler.Re(ctx, consts.Success, nil)
		})
		p.Get("/4", func(ctx iris.Context) {
			handler.Re(ctx, consts.Success, nil)
		})

		p.Get("/5", func(ctx iris.Context) {
			handler.Re(ctx, consts.Success, nil)
		})
		p.Get("/6", func(ctx iris.Context) {
			logic.Send("测试发生消息")
			handler.Re(ctx, consts.Success, nil)
		})
		p.Get("/7", func(ctx iris.Context) {
			handler.ReadingMsg()
			handler.Re(ctx, consts.Success, nil)
		})
		p.Get("/8", handler.Test8)
	})
	app.Get("/test", func(ctx iris.Context) {
		handler.Re(ctx, consts.Success, nil)
	})

	// 任何以/msg为前缀的路径，一律交给websoket处理。
	app.Get("/msg", websocket.Handler(handler.InitWebsocket()))

}
