package models

// UserModel user
type UserModel struct {
	Id       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"-" db:"username"`
	Age      int    `json:"age" db:"age"`
	Sex      string `json:"sex" db:"sex"`
}
