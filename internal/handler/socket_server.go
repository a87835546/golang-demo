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
	"strings"
	"sync"
	"time"
)

type SocketServer struct {
}

var wait = sync.WaitGroup{}
var lo = sync.RWMutex{}

type concurrentMap map[string]*websocket.Conn
type parametersMap map[string]string

var Cmap = concurrentMap{}
var Parameter = parametersMap{}

//SaveParameter 存储请求的参数
func (c parametersMap) SaveParameter(key, value string) {
	lo.Lock()
	defer lo.Unlock()
	c[key] = value
}

//GetParameter 获取请求的参数
func (c parametersMap) GetParameter(key string) string {
	lo.Lock()
	defer lo.Unlock()
	return c[key]
}

// DeleteByWsId 当用户掉线后删除所有的参数列表
func (c parametersMap) DeleteByWsId(id string) {
	for key, _ := range c {
		if strings.Contains(key, id) {
			delete(c, key)
		}
	}
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
		Parameter.DeleteByWsId(c.ID())
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

// UpdateParameter 当前端请求后更新本地维护数据参数
func UpdateParameter(path, uid, param string) {
	key := configKey(path, uid)
	Parameter.SaveParameter(key, param)
}

func SendAll(res any) error {
	for _, conn := range Cmap {
		fmt.Printf("conn id -->>> %s\n", conn.ID())
		result := Result{
			Message: "socket 消息回复" + strconv.FormatInt(time.Now().UnixMilli(), 10),
			Code:    200,
			Data:    res, //"id:" + "--- ws id:" + conn.ID(),
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

func SendOne(uid, path string, f func(param string) (res any)) error {
	// ws 链接对象
	conn := Cmap.GetValue(uid)
	// parameters
	p := Parameter.GetParameter(configKey(path, uid))
	res := f(p)
	result := Result{
		Message: "socket 消息回复" + strconv.FormatInt(time.Now().UnixMilli(), 10),
		Code:    200,
		Data:    res,
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
