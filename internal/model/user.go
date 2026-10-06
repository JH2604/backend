package model

import "time"

type RegisterReq struct {
	StudentID string `json:"student_id" binding:"max=20,min=3,required"`
	Password  string `json:"password" binding:"max=32,min=6,required"`
	Role      string `json:"role" binding:"oneof=student admin,required"`
}

const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StudentID string    `json:"student_id" gorm:"uniqueIndex;size:20"`
	Password  string    `json:"-" gorm:"size:80"`
	CreatedAt time.Time `json:"created_at"`
	Role      string    `json:"role" gorm:"size:20"`
	Name      string    `json:"name" gorm:"size:20"`

	// ---- U1 要返回的资料字段（U2 做完之后才有接口去改它们）----
	AvatarURL   string `json:"avatar_url" gorm:"size:255"` // 头像地址
	Phone       string `json:"phone" gorm:"size:20"`
	Email       string `json:"email" gorm:"size:100"`
	AllowRemind bool   `json:"allow_remind"` // 私信要不要同时发短信/邮件提醒
	Theme       string `json:"theme" gorm:"size:10"` // light / dark / system

	// 不落库：由 service 数出来填进去
	PostCount int64 `gorm:"-" json:"post_count"`
}

type LoginReq struct {
	Userid string `json:"student_id" binding:"max=20,min=3,required"`
	Password string `json:"password" binding:"max=32,min=6,required"`
}

type TokenInfo struct {
	UserID   uint
	Role     string
	SID      string
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PasswordChange struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"max=32,min=6,required"`
}
