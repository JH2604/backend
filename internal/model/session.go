package model

import "time"

type Session struct {
	ID               string     `gorm:"type:char(32);primaryKey" json:"id"`
	UserID           uint       `json:"user_id"`
	AccessTokenHash  string     `gorm:"type:char(64);uniqueIndex" json:"-"`
	RefreshTokenHash string     `gorm:"type:char(64);uniqueIndex" json:"-"`
	PrevRefreshHash  string     `gorm:"type:char(64);Index" json:"-"`
	AccessExpiresAt  time.Time  `json:"access_expires_at"`
	RefreshExpiresAt time.Time  `json:"refresh_expires_at"`
	Platform         string     `gorm:"type:varchar(16)" json:"platform"`
	IP               string     `gorm:"type:varchar(45)" json:"ip"`
	UserAgent        string     `gorm:"type:varchar(255)" json:"user_agent"`
	RevokedAt        *time.Time `json:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at"`
	LastUsedAt       time.Time  `json:"last_used_at"`
}
