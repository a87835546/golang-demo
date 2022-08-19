package middleware

import (
	"fmt"
	"github.com/kataras/iris/v12"
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
