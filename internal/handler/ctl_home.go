package handler

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/logic"
	"golang-demo/internal/models"
)

type HomeCtl struct {
}

// Test 最简单的请求
func (c *HomeCtl) Test(ctx iris.Context) {
	Re(ctx, Success, nil)
}
func (c *HomeCtl) Add(ctx iris.Context) {
	user := models.UserModel{}
	// 获取前端post传递的参数
	ctx.ReadJSON(&user)
	err := logic.UserServiceAddUser(user)
	if err == nil {
		Re(ctx, Success, nil)
	} else {
		Re(ctx, SystemErr, err.Error())
	}
}
func (c *HomeCtl) Query(ctx iris.Context) {
	//获取前端传递的get请求参数
	size := ctx.URLParamIntDefault("size", 10)
	num := ctx.URLParamIntDefault("num", 1)
	fmt.Printf("size --->>> %d num--->>> %d", size, num)
	res, err := logic.UserServiceQueryUser()
	if err == nil {
		Re(ctx, Success, res)
	} else {
		Re(ctx, SystemErr, err.Error())
	}
}
