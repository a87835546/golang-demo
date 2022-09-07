package models

// UserModel user
type UserModel struct {
	//	Base     BaseModel
	Id       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Age      int    `json:"age" db:"age"`
	Sex      string `json:"sex" db:"sex"`
}
