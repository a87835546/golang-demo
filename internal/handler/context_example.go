package handler

import (
	"context"
	"github.com/kataras/iris/v12"
)

type ContextExample struct {
}

func Test(ctx iris.Context) {
	ctx.Values().Get("a")
	context.Background()
	context.TODO()
}
