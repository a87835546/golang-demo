package logic

import (
	"golang-demo/internal/models"
)

type UserService struct {
}

func UserServiceAddUser(ex models.UserModel) error {
	sql, _, err := G.From("user").Insert().Rows(ex).ToSQL()
	_, err = Db.Exec(sql)
	return err
}
