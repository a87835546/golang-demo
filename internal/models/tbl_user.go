package models

// UserModel user
/**
* @author 大菠萝
* @description //TODO 实体结构模型，相当于java语言的实体类
* @date 3:12 pm 9/7/22
* @param 字段熟悉首字母需要大写，否则json跟反射解析不了，
         后面的json是为来跟前端的入参数匹配一般采用小写字母打头。
         db是必须跟数据的字段保持一致，这是goqu的规范
* @return
**/
type UserModel struct {
	//BaseModel
	Id       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Age      int    `json:"age" db:"age"`
	Sex      string `json:"sex" db:"sex"`
}
