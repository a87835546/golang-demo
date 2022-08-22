package logic

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang-demo/internal/consts"
)

var Db *sqlx.DB
var G = goqu.Dialect("mysql")

func init() {
	consts.InitYaml()

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
