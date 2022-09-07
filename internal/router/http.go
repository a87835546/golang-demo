package router

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/kataras/iris/v12/websocket"
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
			handler.Re(ctx, handler.Success, nil)
		})
		p.Get("/2", func(ctx iris.Context) {
			handler.Re(ctx, handler.Success, nil)
		})

		p.Get("/3", func(ctx iris.Context) {
			handler.Re(ctx, handler.Success, nil)
		})
		p.Get("/4", func(ctx iris.Context) {
			handler.Re(ctx, handler.Success, nil)
		})

		p.Get("/5", func(ctx iris.Context) {
			handler.Re(ctx, handler.Success, nil)
		})
		p.Get("/6", func(ctx iris.Context) {
			logic.Send("测试发生消息")
			handler.Re(ctx, handler.Success, nil)
		})
	})
	app.Get("/test", func(ctx iris.Context) {
		handler.Re(ctx, handler.Success, nil)
	})

	app.PartyFunc("/dabluo", func(p router.Party) {
		userCtl := new(handler.UserCtl)
		p.Post("/query", userCtl.QueryUsers)
		p.Post("/queryOne", userCtl.QueryOneUsers)
		p.Post("/addMember", userCtl.AddMember)
		p.Post("/modifyMember", userCtl.ModifyMember)
		p.Post("/deleteMember", userCtl.DeleteMember)
	})

	// websoket 使用
	app.Get("/msg", websocket.Handler(handler.InitWebsocket()))

}
