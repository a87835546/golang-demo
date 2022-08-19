package main

import (
	"errors"
	"fmt"
	"github.com/iris-contrib/swagger/v12"
	"github.com/iris-contrib/swagger/v12/swaggerFiles"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/consts"
	"golang-demo/internal/router"
	"log"
)

func main() {
	fmt.Println("Hello, World!")
	consts.InitYaml()
	app := iris.New()
	url := fmt.Sprintf("http://%v:18081/swagger/doc.json", consts.Conf.Server.Host)
	config1 := &swagger.Config{
		URL: url, //The url pointing to API definition
	}
	log.Printf("url -->>> %s\n", url)
	// use swagger middleware to
	app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(config1, swaggerFiles.Handler))

	router.RouteDemo(app)
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
