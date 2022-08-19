package middleware

import (
	"github.com/kataras/iris/v12"
	"og_agent/doraemon/helper"
	"og_agent/internal/consts"
	"og_agent/internal/handler"
	"og_agent/internal/logic"
	"strings"
	"time"
)

func CheckJWT(ctx iris.Context) {
	tokenString := ctx.GetHeader("token")
	path := ctx.RequestPath(true)
	if strings.Contains(path, "/test/") || strings.Contains(path, "/member/login") {
		ctx.Next()
		return
	}

	if tokenString == "" {
		ctx.StopExecution()
		ctx.StatusCode(iris.StatusUnauthorized)
		handler.Re(ctx, handler.TokenErr, "")
		return
	}
	memberID := logic.GetMemberId(tokenString)
	if memberID != 0 {
		ctx.Values().Set("member_id", memberID)
	}
	//fmt.Printf("token ---->>> %s \nmember id --->>> %d \v", tokenString, memberID)

	wasToken := ctx.Values().GetString("token")
	wasTime := ctx.Values().GetInt64Default("token_ts", 0)
	nowTime := time.Now().Unix()

	if (wasToken == tokenString) && (nowTime-wasTime < int64(time.Second*60)) {
		ctx.Next()
	}

	claims, err := helper.ParseToken(tokenString, consts.JWTSalt)
	if err != nil {
		ctx.StopExecution()
		ctx.StatusCode(iris.StatusUnauthorized)
		handler.Re(ctx, handler.TokenErr, "")
		return
	}

	name := claims.AdminName

	_, ok := ctx.Values().Set("name", name)
	if !ok {
		handler.Re(ctx, 10002, "name")
		return
	}

	_, ok = ctx.Values().Set("token", tokenString)
	if !ok {
		handler.Re(ctx, handler.SystemErr, "")
		return
	}

	_, ok = ctx.Values().Set("token_ts", time.Now().Unix())
	if !ok {
		handler.Re(ctx, handler.SystemErr, "")
		return
	}

	ctx.Next()
}
