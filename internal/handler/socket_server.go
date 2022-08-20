package handler

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/kataras/iris/v12/websocket"
	"github.com/kataras/neffos"
	"github.com/kataras/neffos/gorilla"
	"log"
	"strings"
	"sync"
	"time"
)

type SocketServer struct {
}

var wait = sync.WaitGroup{}
var lo = sync.RWMutex{}

type concurrentMap map[string]*websocket.Conn

var Cmap = concurrentMap{}

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
			pong := strings.Replace(ping, "？", "！", len(ping))
			pong = strings.Replace(pong, "么", "", len(pong))

			result := Result{
				Message: "socket 消息回复" + time.Now().GoString(),
				Code:    200,
				Data:    pong,
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
