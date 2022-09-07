package models

// BaseModel /**
/**
* @author 大菠萝
* @description //TODO 结构体类的基类。每个子查询模型需要继承
* @date 3:19 pm 9/7/22
* @param
         size：查询时每页的数量
         num： 查询时从第num页开始查
* @return
**/
type BaseModel struct {
	Size int    `json:"size" db:"size"`
	Num  string `json:"num" db:"num"`
}
