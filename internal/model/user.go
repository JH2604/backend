package model

import "time"

type RegisterReq struct {
	StudentID string `json:"userid" binding:"max=20,min=3,required"`
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
	Name      string    `json:"name"`
}

type LoginReq struct {
	Userid string `json:"userid" binding:"max=20,min=3,required"`
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
