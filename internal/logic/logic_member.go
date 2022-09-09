package logic

import (
	"fmt"
	"golang-demo/internal/consts"
	"golang-demo/internal/models"
	"golang-demo/internal/repository"
	error_utils "golang-demo/internal/utils/error"
)

// MemberService /**
/**
 * @author 大菠萝
 * @description //TODO 业务逻辑层接口 类似于java的 service的
 * @date 4:13 pm 9/7/22
 * @param
 * @return
 **/
type MemberService interface {
	AddMember(user models.UserModel) error

	QueryMembers(user models.UserModel) ([]models.UserModel, error)

	QueryOneMember(user models.UserModel) (models.UserModel, error)
}

// MemberServiceImpl /**
/**
 * @author 大菠萝
 * @description //TODO MemberService接口的实现结构体
 * @date 4:54 pm 9/7/22
 * @param
 * @return
 **/
type MemberServiceImpl struct {
	repo *repository.UserRepositoryImpl
}

func (memberService *MemberServiceImpl) AddMember(user models.UserModel) error {
	fmt.Printf("进入到业务层查询的user:%v\n", user)
	fmt.Printf("用户名:%s,密码:%s,年龄:%d,性别:%s", user.Username, user.Password, user.Age, user.Sex)
	id, err := memberService.repo.Insert(user)
	fmt.Printf("添加的insert的id:%d", id)
	return err
}

func (memberService *MemberServiceImpl) UpdateMember(user models.UserModel) error {
	fmt.Printf("进入到业务层修改的入参user:%v\n", user)
	fmt.Printf("用户名:%s,密码:%s,年龄:%d,性别:%s", user.Username, user.Password, user.Age, user.Sex)
	id, err := memberService.repo.Update(user)
	fmt.Printf("修改的影响行数:%d", id)
	return err
}

func (memberService *MemberServiceImpl) DeleteMember(user models.UserModel) error {
	fmt.Printf("业务层删除的入参user:%v\n", user)
	fmt.Printf("用户名:%s,密码:%s,年龄:%d,性别:%s", user.Username, user.Password, user.Age, user.Sex)
	id, err := memberService.repo.Delete(user)
	fmt.Printf("删除的影响行数:%d", id)
	return err
}

func (memberService *MemberServiceImpl) QueryMembers(user models.UserModel) ([]models.UserModel, error) {
	//userSlices := make([]models.UserModel, 0)
	fmt.Printf("进入到业务层查询的user:%v\n", user)
	userSlices, err := memberService.repo.SelectMembers(user)
	if true {
		panic(error_utils.ServiceErrorModel{Code: consts.TokenErr})
	}
	return userSlices, err
}

func (memberService *MemberServiceImpl) QueryOneMember(user models.UserModel) (models.UserModel, error) {
	//userSlices := make([]models.UserModel, 0)
	fmt.Printf("进入到业务层查询的user:%v\n", user)
	userModel, err := memberService.repo.SelectOneMember(user)
	return userModel, err
}
