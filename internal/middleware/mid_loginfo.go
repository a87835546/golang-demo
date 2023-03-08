package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kataras/iris/v12"
	"log"
	"time"
)

func LogInfoBefore(ctx iris.Context) {

	timeBegin := time.Now()

	ctx.Values().Set("tn", timeBegin.UnixNano())
	path := ctx.RequestPath(true)

	//body := ctx.Request().Body
	//ps, _ := ioutil.ReadAll(body)
	//
	//ctx.Request().Body. = bytes.NewReader()

	fmt.Printf(" ----->>>> [%s]", path)

	ctx.Next()
}

func LogInfoBefore1() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		if c.Request.URL.Path == "/foo" {
			return
		}
		// 可以通过上下文对象，设置一些依附在上下文对象里面的键/值数据
		c.Set("example", "12345")

		// 在这里处理请求到达控制器函数之前的逻辑

		// 调用下一个中间件，或者控制器处理函数，具体得看注册了多少个中间件。
		c.Next()

		// 在这里可以处理请求返回给用户之前的逻辑
		latency := time.Since(t)
		log.Print(latency)

		// 例如，查询请求状态吗
		status := c.Writer.Status()
		log.Println(status)
	}
}

func LogInfoAfter(ctx iris.Context) {
	timeBeginNano := ctx.Values().Get("tn").(int64)
	latency := float64(time.Now().UnixNano()-timeBeginNano) / 1000000000

	// latency := time.Since()
	//rs := ctx.Values().Get("data").(string)

	//timeFormat := time.Unix(0, timeBeginNano).Format("15:04:[5.00000] 01-02")
	//global.Logger.Infof("%s <-- %f %s", timeFormat, latency, rs)
	fmt.Printf(" <<<<------- %f \n", latency)
	ctx.Next()
}
