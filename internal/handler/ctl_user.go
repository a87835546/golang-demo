package handler

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/consts"
	"golang-demo/internal/logic"
	"golang-demo/internal/models"
)

// UserCtl /** fat分支修改
/**
 * @author 大菠萝
 * @description //TODO 控制层结构体 类似于 java中的control
 * @date 4:14 pm 9/7/22
 * @param
 * @return
 **/
type UserCtl struct {
	Service logic.MemberServiceImpl
}

// QueryUsers /**
/**
 * @author 大菠萝
 * @description //TODO 控制层中查询用户列表的接口
 * @date 4:15 pm 9/7/22
 * @param //TODO iris的上下文对象，可以通过该上下文对象获取前端的传过来的相关参数及请求头信息
 * @return
 **/
func (c *UserCtl) QueryUsers(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	defer HandlePanic(ctx, nil)
	user := models.UserModel{}
	//TODO 从上下文对象中获取json参数并组装到对应的user结构体数据中
	ctx.ReadJSON(&user)
	fmt.Printf("查询列表参数%+v", user)
	fmt.Println("用户名称" + user.Username)
	userSlices, err := (&c.Service).QueryMembers(user)
	fmt.Printf("获取到的切片数据:%v\n", userSlices)
	if err != nil {
		HandleErr(ctx, nil, err)
		return
	}
	Re(ctx, consts.Success, userSlices)
}

func (c *UserCtl) QueryOneUsers(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	userOne, _ := (&c.Service).QueryOneMember(user)
	fmt.Printf("获取到单条数据:%v\n", userOne)
	Re(ctx, consts.Success, userOne)
}

func (c *UserCtl) AddMember(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).AddMember(user)
	fmt.Printf("添加结束后的数据:%v\n", err)
	Re(ctx, consts.Success, nil)
}

func (c *UserCtl) ModifyMember(ctx iris.Context) {
	fmt.Println("我是大菠萝修改")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).UpdateMember(user)
	fmt.Printf("修改结束后的数据:%v\n", err)
	Re(ctx, consts.Success, nil)
}

func (c *UserCtl) DeleteMember(ctx iris.Context) {
	fmt.Println("我是大菠萝删除")
	user := models.UserModel{}
	ctx.ReadJSON(&user)
	fmt.Println("用户名称" + user.Username)
	err := (&c.Service).DeleteMember(user)
	fmt.Printf("删除结束后的数据:%v\n", err)
	Re(ctx, consts.Success, nil)
}

func (c *UserCtl) QueryUsersByPages(ctx iris.Context) {
	fmt.Println("我是大菠萝")
	defer HandlePanic(ctx, nil)
	user := models.UserModel{}
	//TODO 从上下文对象中获取json参数并组装到对应的user结构体数据中
	ctx.ReadJSON(&user)
	fmt.Printf("查询列表参数%+v", user)
	fmt.Println("用户名称" + user.Username)
	vo, err := (&c.Service).QueryMembersByPage(user)
	fmt.Printf("获取到的切片数据:%v\n", vo)
	if err != nil {
		HandleErr(ctx, nil, err)
		return
	}
	Re(ctx, consts.Success, vo)
}
