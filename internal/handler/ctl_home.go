package handler

import "github.com/kataras/iris/v12"

type HomeCtl struct {
}

// Test 最简单的请求
func (c *HomeCtl) Test(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"message": "ceshi",
		"code":    200,
	})
	ctx.Next()
}
