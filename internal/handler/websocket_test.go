package handler

import (
	"testing"
)

func TestSendAll(t *testing.T) {
	SendAll("查询到的所有参数")
}

func TestParametersMap_Clear(t *testing.T) {
	SendOne("id", "path", func(param string) (res any) {
		//param 的参数去获取知道数据 param 需要转换成 map
		return "使用param 查询到的所有参数"
	})
}
