package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"golang-demo/internal/models"
)

var SqlxDB *sqlx.DB

type UserRepository interface {
	SelectMembers(user models.UserModel) ([]models.UserModel, error)
}

type UserRepositoryImpl struct {
}

func (userRepositoryImpl *UserRepositoryImpl) SelectMembers(user models.UserModel) ([]models.UserModel, error) {
	sql := "select id,username,`password`,age,sex from user where username = ?"
	//slicesUser := make([]models.UserModel, 0)
	var slicesUser []models.UserModel
	err := SqlxDB.Select(&slicesUser, sql, user.Username)
	fmt.Printf("查询数据库入参:%s,返回的数值:%v,%v", user.Username, slicesUser, err)
	return slicesUser, err

}

func (userRepositoryImpl *UserRepositoryImpl) SelectOneMember(user models.UserModel) (models.UserModel, error) {
	sql := "select id,username,`password`,age,sex from user where id = ?"
	var userOne models.UserModel
	err := SqlxDB.Get(&userOne, sql, user.Id)
	fmt.Printf("查询数据库单条入参:%d,返回的数值:%v,%v", user.Id, userOne, err)
	return userOne, err
}

func (userRepositoryImpl *UserRepositoryImpl) Insert(user models.UserModel) (int64, error) {
	sql := "INSERT INTO user (username,`password`,age,sex) VALUES (?, ?, ?,?)"
	result, err := SqlxDB.Exec(sql, user.Username, user.Password, user.Age, user.Sex)
	fmt.Printf("查询数据库单条入参:%v,返回的数值:%v,%v", user, result, err)
	id, err2 := result.LastInsertId()
	return id, err2
}

func (userRepositoryImpl *UserRepositoryImpl) Update(user models.UserModel) (int64, error) {
	sql := "update user set age = ? where id = ?"
	result, err := SqlxDB.Exec(sql, user.Age, user.Id)
	fmt.Printf("修改数据库单条入参:%v,返回的数值:%v,%v", user, result, err)
	affectNum, err2 := result.RowsAffected()
	return affectNum, err2
}

func (userRepositoryImpl *UserRepositoryImpl) Delete(user models.UserModel) (int64, error) {
	sql := "DELETE from user where id = ?"
	result, err := SqlxDB.Exec(sql, user.Id)
	fmt.Printf("删除数据库入参:%v,返回的数值:%v,%v", user, result, err)
	affectNum, err2 := result.RowsAffected()
	return affectNum, err2
}
