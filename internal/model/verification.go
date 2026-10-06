package model

import "time"

type VerificationCode struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Target    string     `json:"target" gorm:"index:idx_target_scene;size:100"`
	Code      string     `json:"code" gorm:"size:6"`
	Scene     string     `json:"scene" gorm:"index:idx_target_scene;size:20"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type SendCodeReq  struct {
	Channel string `json:"channel"   binding:"required,oneof=sms email"`
	Target  string `json:"target"    binding:"required"`
	Scene   string `json:"scene"     binding:"required"`
}

type BindContactReq struct{
	Channel string  `json:"channel"   binding:"required,oneof=sms email"`
	Target  string  `json:"target"    binding:"required"`
	Code    string  `json:"code"      binding:"required"`

}