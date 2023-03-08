package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iris-contrib/swagger/v12"
	"github.com/iris-contrib/swagger/v12/swaggerFiles"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/consts"
	"golang-demo/internal/middleware"
	"golang-demo/internal/router"
	"log"
)

func main() {
	fmt.Println("Hello, World!")

	g := gin.Default()
	g.Use(middleware.LogInfoBefore1(), gin.Logger(), gin.Recovery())
	router.RouteDemoUsingGin(g)

	//TODO 初始化配置常量信息到内存对象中
	consts.InitYaml()

	//TODO 初始化iris的上下文对象
	app := iris.New()

	//TODO 初始化swagger的连接信息
	url := fmt.Sprintf("http://%v:18081/swagger/doc.json", consts.Conf.Server.Host)
	config1 := &swagger.Config{
		URL: url, //The url pointing to API definition
	}
	log.Printf("url -->>> %s\n", url)
	// use swagger middleware to
	app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(config1, swaggerFiles.Handler))

	//TODO 通过iris框架对 http的请求路径跟handler进行路由映射
	router.RouteDemo(app)

	/*** 启动服务 ***/
	serverConf := fmt.Sprintf("%s:%d", consts.Conf.Server.Host, consts.Conf.Server.Port)
	err := app.Run(
		iris.Addr(serverConf),                         // 启动服务，监控地址127.0.0.1和端口18082
		iris.WithoutServerError(iris.ErrServerClosed), // 忽略服务器错误
		iris.WithOptimizations,                        // 让程序自身尽可能的优化
		iris.WithCharset("UTF-8"),                     // 国际化
	)

	if err != nil {
		panic(errors.New("######### project service STAR FAILED with HttpService #########"))
	}

	//TODO 异常捕获
	defer CatchPanic()
}
func CatchPanic() {
	if r := recover(); r != nil {
		fmt.Println("Panic", r)
	}
}
