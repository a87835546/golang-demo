package handler

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris/v12"
)

// Result 返回数据结构/
type Result struct {
	Message string `json:"message"`
	Code    int32  `json:"code"`
	Data    any    `json:"data"`
}

// PaginationResult 带分页的数据模型
type PaginationResult struct {
	Total int `json:"total"`
	Size  int `json:"size"`
	Page  int `json:"page"`
	Data  any `json:"data"`
}

type BasicCtl struct{}

type JsonResult struct {
	Code int32       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func jsonSerialize(c interface{}) string {
	data, err := json.Marshal(c) //序列化，返回data为bytes类型
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s", data)
}

func genJsonString(key, value string) string {

	bs, _ := json.Marshal(map[string]string{key: value})

	return string(bs)
}

// ResultSuccess 构建返回成功的结构
func ResultSuccess(data any) (res Result) {
	return Result{
		"request success",
		200,
		data,
	}
}

// Re 通用 response 返回封装
func Re(ctx iris.Context, errCode int32, data interface{}) {
	rzt := Result{
		Code:    errCode,
		Message: MessageMap[errCode],
		Data:    data,
	}
	//TODO 通过上下文对象把查询到的返回值按统一的数据格式返回给前端
	_, _ = ctx.JSON(rzt)

	ctx.Values().Set("data", jsonSerialize(rzt))
	ctx.Next()
}
