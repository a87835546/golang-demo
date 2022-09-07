package handler

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/logic"
	"golang-demo/internal/models"
)

type UserCtl struct {
	Service logic.MemberServiceImpl
}

func (c *UserCtl) QueryUsers(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Printf("查询列表参数%+v", user)
	fmt.Println("用户名称" + user.Username)
	userSlices, _ := (&c.Service).QueryMembers(user)
	fmt.Printf("获取到的切片数据:%v\n", userSlices)
	Re(ctx, Success, userSlices)
}

func (c *UserCtl) QueryOneUsers(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	userOne, _ := (&c.Service).QueryOneMember(user)
	fmt.Printf("获取到单条数据:%v\n", userOne)
	Re(ctx, Success, userOne)
}

func (c *UserCtl) AddMember(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).AddMember(user)
	fmt.Printf("添加结束后的数据:%v\n", err)
	Re(ctx, Success, nil)
}

func (c *UserCtl) ModifyMember(ctx iris.Context) {
	fmt.Println("我是大菠萝修改")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).UpdateMember(user)
	fmt.Printf("修改结束后的数据:%v\n", err)
	Re(ctx, Success, nil)
}

func (c *UserCtl) DeleteMember(ctx iris.Context) {
	fmt.Println("我是大菠萝删除")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).DeleteMember(user)
	fmt.Printf("删除结束后的数据:%v\n", err)
	Re(ctx, Success, nil)
}
