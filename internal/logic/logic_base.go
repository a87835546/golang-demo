package logic

import (
	"fmt"
	"github.com/beanstalkd/go-beanstalk"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang-demo/internal/consts"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

var Db *sqlx.DB
var G = goqu.Dialect("mysql")
var BeanstalkdConn = &beanstalk.Conn{}

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Dial(addr string) (net.Conn, error) {
	return self.client.Dial("tcp", addr)
}
func init() {
	consts.InitYaml()
	InitBeanstalkd()
	_dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/demo?charset=utf8mb4&parseTime=True", consts.Conf.MySQL.User, consts.Conf.MySQL.Password, consts.Conf.MySQL.Host)
	//_dsn := "admin:qwer1234@tcp(test.cjyntu0au13f.ap-southeast-1.rds.amazonaws.com:3306)/demo?charset=utf8mb4&parseTime=True"
	td, err := sqlx.Connect("mysql", _dsn)
	if nil != err {
		fmt.Printf("connect DB failed,err:%v\n", err)
		return
	}
	Db = td
	fmt.Println("db", Db)
	Db.SetMaxOpenConns(20)
	Db.SetMaxIdleConns(10)
}

func InitBeanstalkd() {
	addr := fmt.Sprintf("%s:%d", consts.Conf.Beanstalkd.Host, consts.Conf.Beanstalkd.Port)
	fmt.Printf("add -->> %s", addr)
	c, err := beanstalk.Dial("tcp", addr)
	BeanstalkdConn = c
	if err != nil {
		fmt.Printf("coconnect beanstalkd err -->>> %s", err.Error())
	}
}

// Send 生产消息
func Send(msg string) {
	id, err := BeanstalkdConn.Put([]byte(msg), 1, 0, 120*time.Second)
	if err != nil {
		fmt.Printf("sending beanstalkd err -->>> %s", err.Error())
	} else {
		fmt.Printf("id --- >>> %v", id)
	}
}
