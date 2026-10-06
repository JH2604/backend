package model

import "time"

// U6 / U7 的返回结构（DTO）——不能直接返回 model.User，
// 否则普通用户会拿到 StudentID / Phone / Email 等隐私字段

// AdminBrief U6 管理员列表里的项
type AdminBrief struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	// CanMessage 目前没有黑名单概念，恒为 true（计算字段，不落库）
	CanMessage bool `json:"can_message"`
}

// UserDetail U7 里只有管理员能看到的隐私字段
// 由 UserPublic.Detail 这个指针控制整体是否输出 null，所以内部不用 omitempty
type UserDetail struct {
	StudentID   string     `json:"student_id"`
	Phone       string     `json:"phone"`
	Email       string     `json:"email"`
	AllowRemind bool       `json:"allow_remind"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// UserPublic U7 查看发帖人信息的返回体
type UserPublic struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	PostCount int64  `json:"post_count"`
	// CanMessage 同 U6：暂无黑名单，恒 true
	CanMessage bool `json:"can_message"`
	// CanRemind 规则见 service.canRemind()，与 M4 的 can_remind 共用
	CanRemind bool `json:"can_remind"`
	// 普通用户看不到 → 保持 nil → JSON 里是 null
	Detail *UserDetail `json:"detail"`
}
