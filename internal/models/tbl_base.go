package models

type BaseModel struct {
	Size int    `json:"size" db:"size"`
	Num  string `json:"num" db:"num"`
}
