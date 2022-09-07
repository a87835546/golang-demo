package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"golang-demo/internal/models"
)

var SqlxDB *sqlx.DB

// UserRepository /**
/**
 * @author 大菠萝
 * @description //TODO 数据库user表持久化接口
 * @date 3:18 pm 9/7/22
 * @param
 * @return
 **/
type UserRepository interface {
	SelectMembers(user models.UserModel) ([]models.UserModel, error)
	SelectOneMember(user models.UserModel) (models.UserModel, error)
	Insert(user models.UserModel) (int64, error)
	Update(user models.UserModel) (int64, error)
	Delete(user models.UserModel) (int64, error)
}

// UserRepositoryImpl /**
/**
 * @author 大菠萝
 * @description //TODO UserRepository接口的实现结构体
 * @date 3:25 pm 9/7/22
 * @param
 * @return
 **/
type UserRepositoryImpl struct {
}

// SelectMembers /**
/**
 * @author 大菠萝
 * @description //TODO 查询会员列表
 * @date 3:27 pm 9/7/22
 * @param
 * @return 会员列表切片数据
 **/
func (userRepositoryImpl *UserRepositoryImpl) SelectMembers(user models.UserModel) ([]models.UserModel, error) {
	sql := "select id,username,`password`,age,sex from user where username = ?"
	//slicesUser := make([]models.UserModel, 0)
	var slicesUser []models.UserModel
	err := SqlxDB.Select(&slicesUser, sql, user.Username)
	fmt.Printf("查询数据库入参:%s,返回的数值:%v,%v", user.Username, slicesUser, err)
	return slicesUser, err

}

// SelectOneMember /**
/**
 * @author 大菠萝
 * @description //TODO 查询单个用户方法
 * @date 3:30 pm 9/7/22
 * @param
 * @return 单个会员记录
 **/
func (userRepositoryImpl *UserRepositoryImpl) SelectOneMember(user models.UserModel) (models.UserModel, error) {
	sql := "select id,username,`password`,age,sex from user where id = ?"
	var userOne models.UserModel
	err := SqlxDB.Get(&userOne, sql, user.Id)
	fmt.Printf("查询数据库单条入参:%d,返回的数值:%v,%v", user.Id, userOne, err)
	return userOne, err
}

// Insert /**
/**
 * @author 大菠萝
 * @description TODO 向user表插入一条记录
 * @date 3:30 pm 9/7/22
 * @param
 * @return TODO id:主键编号
 **/
func (userRepositoryImpl *UserRepositoryImpl) Insert(user models.UserModel) (int64, error) {
	sql := "INSERT INTO user (username,`password`,age,sex) VALUES (?, ?, ?,?)"
	result, err := SqlxDB.Exec(sql, user.Username, user.Password, user.Age, user.Sex)
	fmt.Printf("查询数据库单条入参:%v,返回的数值:%v,%v", user, result, err)
	id, err2 := result.LastInsertId()
	return id, err2
}

// Update /**
/**
 * @author 大菠萝
 * @description //TODO 修改方法
 * @date 3:31 pm 9/7/22
 * @param
 * @return //TODO affectNum：修改的条数
 **/
func (userRepositoryImpl *UserRepositoryImpl) Update(user models.UserModel) (int64, error) {
	sql := "update user set age = ? where id = ?"
	result, err := SqlxDB.Exec(sql, user.Age, user.Id)
	fmt.Printf("修改数据库单条入参:%v,返回的数值:%v,%v", user, result, err)
	affectNum, err2 := result.RowsAffected()
	return affectNum, err2
}

// Delete /**
/**
 * @author 大菠萝
 * @description //TODO 删除的方法
 * @date 3:32 pm 9/7/22
 * @param
 * @return //TODO affectNum：删除的条数
 **/
func (userRepositoryImpl *UserRepositoryImpl) Delete(user models.UserModel) (int64, error) {
	sql := "DELETE from user where id = ?"
	result, err := SqlxDB.Exec(sql, user.Id)
	fmt.Printf("删除数据库入参:%v,返回的数值:%v,%v", user, result, err)
	affectNum, err2 := result.RowsAffected()
	return affectNum, err2
}
