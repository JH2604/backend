package model

import "time"

// ============ U6 / U7 的返回结构（DTO）============
// ⚠️ 为什么不直接 response.Success(c, user) 把 model.User 丢出去？
// 因为 User 带着 StudentID / Phone / Email 这些隐私字段，
// 而「看别人资料」的接口不能让普通用户拿到它们。
// Password 已经用 json:"-" 挡住了，但学号/手机/邮箱没挡，
// 所以这两个接口必须走专门的 DTO。

// AdminBrief：U6「管理员列表」里的一项
type AdminBrief struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	//AvatarURL头像图片的网址
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"` // 恒为 admin
	Email     string `json:"email"`
	// CanMessage：计算字段，不落库：能不能给他发私信
	CanMessage bool `json:"can_message"`
}

// UserDetail：U7 里「只有管理员能看到」的隐私字段
// 整体被 UserPublic.Detail 这个指针控制是不是 null，所以内部不用 omitempty
type UserDetail struct {
	StudentID string `json:"student_id"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	//AllowRemind：他本人是否愿意收提醒
	AllowRemind bool `json:"allow_remind"`
	//CreatedAt注册时间
	CreatedAt time.Time `json:"created_at"`
	//LastLoginAt最近登陆
	LastLoginAt *time.Time `json:"last_login_at"`
}

// UserPublic：U7「查看发帖人信息」的返回体
type UserPublic struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	//PostCount：他发过多少条帖
	PostCount int64 `json:"post_count"`
	//`CanMessage` ← 业务规则：目前没有黑名单概念 → 恒 `true`
	CanMessage bool `json:"can_message"`
	//`CanRemind` ← 业务规则：`对方开了提醒开关 && (绑了手机 || 绑了邮箱`）
	CanRemind bool `json:"can_remind"`
	// 普通用户看不到 → 保持 nil → JSON 里就是 null
	Detail *UserDetail `json:"detail"`
}
