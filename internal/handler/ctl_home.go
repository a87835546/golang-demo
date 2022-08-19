package handler

import "github.com/kataras/iris/v12"

type HomeCtl struct {
}

// Test 最简单的请求
func (c *HomeCtl) Test(ctx iris.Context) {
	Re(ctx, Success, nil)
}
