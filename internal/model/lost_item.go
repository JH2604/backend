package model

// 1. 数据库模型：多了 ID 字段
type LostItem struct {
	//uint为无符号整数，即没有负数的int，且数据范围是int的2倍
	//gorm:"primaryKey":告诉gorm，id是主键，类比数组的每个元素唯一下标
	ID uint `gorm:"primaryKey" json:"id"` // GORM 默认认为 ID 是主键，自增

	//binding:"required":它告诉 Gin 框架：“当前端发来 JSON 时，这个字段绝对不能为空，如果没有，直接报错拦截！”
	Title    string `json:"title" binding:"required"`
	Location string `json:"location"`
	Desc     string `json:"desc"`
}
