package handler

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos"
	"golang-demo/doraemon/gorilla"
	"log"
	"strconv"
	"sync"
	"time"
)

type SocketServer struct {
}

var wait = sync.WaitGroup{}
var lo = sync.RWMutex{}

type concurrentMap map[string]*websocket.Conn
type parametersMap map[string]any

var Cmap = concurrentMap{}
var Parameter = parametersMap{}

//SaveParameter 存储请求的参数
func (c parametersMap) SaveParameter(key string, value any) {
	lo.Lock()
	defer lo.Unlock()
	c[key] = value
}

//GetParameter 获取请求的参数
func (c parametersMap) GetParameter(key string) any {
	lo.Lock()
	defer lo.Unlock()
	return c[key]
}

func (c parametersMap) Delete(key string) {
	lo.Lock()
	lo.Unlock()
	delete(c, key)
}
func (c concurrentMap) setValue(key string, value *websocket.Conn) {
	lo.Lock()
	defer lo.Unlock()
	c[key] = value
}

func (c concurrentMap) GetValue(key string) *websocket.Conn {
	lo.Lock()
	defer lo.Unlock()
	return c[key]
}

func InitWebsocket() *neffos.Server {
	ws := websocket.New(gorilla.DefaultUpgrader, websocket.Events{
		websocket.OnNativeMessage: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("Server got: %s from [%s] [%s]", msg.Body, nsConn.Conn.ID(), nsConn.Conn)

			ping := string(msg.Body)

			result := Result{
				Message: "socket 消息回复" + time.Now().GoString() + nsConn.Conn.ID(),
				Code:    200,
				Data:    ping,
			}
			log.Printf("res --->>>> %v\n", result)
			mg := websocket.Message{
				Body:     result.ToBytes(),
				IsNative: true,
			}

			nsConn.Conn.Write(mg)
			return nil
		},
	})

	ws.OnConnect = func(c *websocket.Conn) error {
		log.Printf("[%s] Connected to server!", c.ID())
		if _, ok := Cmap[c.ID()]; ok != true {
			Cmap.setValue(c.ID(), c)
		}
		return nil
	}

	ws.OnDisconnect = func(c *websocket.Conn) {
		log.Printf("[%s] Disconnected from server", c.ID())
		delete(Cmap, c.ID())
		delete(Parameter, c.ID())
	}

	ws.OnUpgradeError = func(err error) {
		log.Printf("Upgrade Error: %v", err)
	}
	return ws
}
func (r Result) ToBytes() []byte {
	buf := new(bytes.Buffer)
	if s, err := json.Marshal(&r); err != nil {
		log.Printf("解析数据异常--->>>%v \n", s)
		return nil
	} else {
		if err := binary.Write(buf, binary.BigEndian, s); err != nil {
			log.Printf("err --->>>> %v \n", err.Error())
			return nil
		}
		return buf.Bytes()
	}
}

func configKey(path, uid string) string {
	conn := Cmap.GetValue(uid)
	key := fmt.Sprintf("id:%s,path:%s,wsid:%s", uid, path, conn.ID())
	return key
}

func UpdateParameter(path, uid string, param any) {
	key := configKey(path, uid)
	Parameter.SaveParameter(key, param)
}

func Send(path, uid, msg string, f func(key string)) error {
	conn := Cmap.GetValue(uid)
	key := configKey(path, uid)
	f(key)
	result := Result{
		Message: "socket 消息回复" + strconv.FormatInt(time.Now().UnixMilli(), 10),
		Code:    200,
		Data:    "id:" + msg + "--- ws id:" + conn.ID(),
	}
	mg := websocket.Message{
		Body:     result.ToBytes(),
		IsNative: true,
	}
	log.Printf("res --->>>> %v\n", result)

	if ok := conn.Write(mg); ok {
		return nil
	} else {
		return errors.New("send error")
	}
}

func SendAll() error {
	for _, conn := range Cmap {
		fmt.Printf("conn id -->>> %s\n", conn.ID())
		result := Result{
			Message: "socket 消息回复" + strconv.FormatInt(time.Now().UnixMilli(), 10),
			Code:    200,
			Data:    "id:" + "--- ws id:" + conn.ID(),
		}
		mg := websocket.Message{
			Body:     result.ToBytes(),
			IsNative: true,
		}
		log.Printf("res --->>>> %v\n", result)

		if ok := conn.Write(mg); ok {
			return nil
		} else {
			return errors.New("send error")
		}
	}
	return nil
}

func SendOne(id string) error {
	conn := Cmap.GetValue(id)
	result := Result{
		Message: "socket 消息回复" + strconv.FormatInt(time.Now().UnixMilli(), 10),
		Code:    200,
		Data:    "id:" + "--- ws id:" + conn.ID(),
	}
	mg := websocket.Message{
		Body:     result.ToBytes(),
		IsNative: true,
	}
	log.Printf("res --->>>> %v\n", result)

	if ok := conn.Write(mg); ok {
		return nil
	} else {
		return errors.New("send error")
	}
}
