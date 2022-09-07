package middleware

import "github.com/kataras/iris/v12"

// CROS /**
/**
 * @author 大菠萝
 * @description //TODO iris跨越的处理
 * @date 3:41 pm 9/7/22
 * @param iris上下文环境对象
 * @return
 **/
func CROS(ctx iris.Context) {

	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	if ctx.Method() == iris.MethodOptions {
		ctx.Header("Access-Control-Methods", "POST, PUT, PATCH, DELETE")
		ctx.Header("Access-Control-Allow-Headers", "Access-Control-Allow-Origin,Content-Type,X-API-CHANNEL,Token")
		ctx.Header("Access-Control-Max-Age", "86400")
		ctx.StatusCode(iris.StatusNoContent)
		return
	}
	ctx.Next()
}
