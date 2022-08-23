package logic

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"golang-demo/internal/models"
)

type UserService struct {
}

func UserServiceAddUser(ex models.UserModel) error {
	sql, _, err := G.From("user").Insert().Rows(ex).ToSQL()
	_, err = Db.Exec(sql)
	return err
}
func UserServiceUpdateUser(user models.UserModel) error {
	rd := goqu.Ex{"name": "lisi"}
	sql, _, err := G.From("user").Update().Where(goqu.Ex{"id": 1}).Set(rd).ToSQL()
	_, err = Db.Exec(sql)

	return err
}

func UserServiceQueryUser() (v []models.UserModel, err error) {
	sql, _, err := G.From("user").Select().Where(goqu.Ex{"id": 1}).Limit(10).Offset(1).ToSQL()
	fmt.Printf("sql --->>> %s \n", sql)
	sql, _, err = G.From("user").Select().ToSQL()
	fmt.Printf("sql --->>> %s \n", sql)
	res, err := Db.Queryx(sql)
	for res.Next() {
		var p models.UserModel
		err = res.StructScan(&p)
		v = append(v, p)
	}
	return v, err
}
