package model

import "time"

type RegisterReq struct {
	Username string `json:"username" binding:"max=20,min=3,required"`
	Password string `json:"password" binding:"max=32,min=6,required"`
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `json:"username" binding:"required" gorm:"uniqueIndex;size:20"`
	Password  string    `json:"password" gorm:"size:80"`
	CreatedAt time.Time `json:"created_at"`
}
