package main

import (
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/consts"
	"golang-demo/internal/router"
)

func main() {
	fmt.Println("Hello, World!")
	app := iris.New()
	router.RouteDemo(app)
	consts.InitYaml()
	/*** 启动服务 ***/
	serverConf := fmt.Sprintf("%s:%d", consts.Conf.Server.Host, consts.Conf.Server.Port)
	err := app.Run(
		iris.Addr(serverConf),                         // 启动服务，监控地址和端口
		iris.WithoutServerError(iris.ErrServerClosed), // 忽略服务器错误
		iris.WithOptimizations,                        // 让程序自身尽可能的优化
		iris.WithCharset("UTF-8"),                     // 国际化
	)

	if err != nil {
		panic(errors.New("######### project service STAR FAILED with HttpService #########"))
	}
	defer CatchPanic()
}
func CatchPanic() {
	if r := recover(); r != nil {
		fmt.Println("Panic", r)
	}
}
