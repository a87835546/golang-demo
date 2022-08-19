package handler

import (
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
	ctx.ReadJSON(&user)
	err := logic.UserServiceAddUser(user)
	if err == nil {
		Re(ctx, Success, nil)
	} else {
		Re(ctx, SystemErr, err.Error())
	}
}
