package handler

import (
	"testing"
)

//func TestWebsocket(t *testing.T) {
//	result := Result{
//		Message: "socket 消息回复" + time.Now().GoString(),
//		Code:    200,
//		Data:    "ping",
//	}
//	log.Printf("res --->>>> %v\n", result)
//	mg := websocket.Message{
//		Body:     result.ToBytes(),
//		IsNative: true,
//	}
//	WSConn.Write(mg)
//}

func TestSendAll(t *testing.T) {
	SendAll()
}

//func TestParametersMap_Clear(t *testing.T) {
//
//	for i := 0; i < 10000; i++ {
//		Parameter.SaveParameter(fmt.Sprintf("time:%d-%d", time.Now().UnixNano(), i), i)
//	}
//	fmt.Printf("param -->> %s", Parameter)
//
//	Parameter.Clear()
//	fmt.Printf("param -->> %s", Parameter)
//}
