package logic

import (
	"fmt"
	"github.com/beanstalkd/go-beanstalk"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang-demo/internal/consts"
	"golang-demo/internal/repository"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

// Db G BeanstalkdConn/**
/**
 * @author 大菠萝
 * @description //TODO 定义三个全局变量分别是 sqlx的连接对象、goqu的连接对象、beanstalk的的连接对象
 * @date 3:56 pm 9/7/22
 * @param
 * @return
 **/
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
	//TODO 初始化配置信息，把proj-dev.yaml文件里的配置信息存放到consts.Conf的结构体里。
	//TODO 配置信息如下：http服务ip跟端口号，mysql的服务信息
	consts.InitYaml()

	//TODO 初始化Beanstalkd连接
	InitBeanstalkd()
	_dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/demo?charset=utf8mb4&parseTime=True", consts.Conf.MySQL.User, consts.Conf.MySQL.Password, consts.Conf.MySQL.Host)
	//_dsn := "admin:qwer1234@tcp(test.cjyntu0au13f.ap-southeast-1.rds.amazonaws.com:3306)/demo?charset=utf8mb4&parseTime=True"

	//TODO 创建sqlx的连接实列
	td, err := sqlx.Connect("mysql", _dsn)
	if nil != err {
		fmt.Printf("connect DB failed,err:%v\n", err)
		return
	}
	//TODO 对全局sqlx的连接对象指针初始化，否则会出现空指针异常
	Db = td
	fmt.Println("db", Db)
	//TODO 配置连接池最大连接数
	Db.SetMaxOpenConns(20)
	//TODO 配置连接池最大核型数
	Db.SetMaxIdleConns(10)
	repository.SqlxDB = Db
}

// InitBeanstalkd /**
/**
 * @author 大菠萝
 * @description //TODO 初始化Beanstalkd生产者的连接
 * @date 4:44 pm 9/7/22
 * @param
 * @return
 **/
func InitBeanstalkd() {
	addr := fmt.Sprintf("%s:%d", consts.Conf.Beanstalkd.Host, consts.Conf.Beanstalkd.Port)
	fmt.Printf("add -->> %s", addr)
	//t := beanstalk.Tube{}
	c, err := beanstalk.Dial("tcp", addr)
	BeanstalkdConn = c

	if err != nil {
		fmt.Printf("coconnect beanstalkd err -->>> %s \n", err.Error())
		fmt.Printf("coconnect beanstalkd err -->>> %v \n", c)
	}
}

// Send /**
/**
 * @author 大菠萝
 * @description //TODO 生产者向中间件发送消息
 * @date 4:06 pm 9/7/22
 * @param Beanstalkd 发送的字符串消息
 * @return
 **/
func Send(msg string) {
	id, err := BeanstalkdConn.Put([]byte(msg), 1, 0, 120*time.Second)
	if err != nil {
		fmt.Printf("sending beanstalkd err -->>> %s", err.Error())
	} else {
		fmt.Printf("id --- >>> %v", id)
	}
}
