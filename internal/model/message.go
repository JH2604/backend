package model

import (
	"time"

	"gorm.io/gorm"
)

// Message：私信实体，对应数据库表 messages
// 设计覆盖 M1~M5 全部需求，一次设计好，避免以后反复改表
type Message struct {
	// 主键：自增 ID
	ID uint `gorm:"primaryKey" json:"id"`

	// 发送者 ID → users.id
	// index：M2 里"我发出的消息"会按这个字段查，加索引提速
	SenderID uint `gorm:"index;not null" json:"sender_id"`

	// 接收者 ID → users.id
	// 用【复合索引】：M1 的查询是 receiver_id + is_read，
	// 两个字段建在同一个索引 idx_receiver_read 上，查询最快
	ReceiverID uint `gorm:"index:idx_receiver_read;not null" json:"receiver_id"`

	// 私信内容
	Content string `gorm:"type:text;not null" json:"content"`

	// 是否已读。默认 false（刚发出来就是未读）
	// 和 ReceiverID 共用复合索引，服务于 M1 的"未读计数"
	IsRead bool `gorm:"index:idx_receiver_read;not null;default:false" json:"is_read"`

	// 已读时间。未读时为 nil，序列化成 null
	ReadAt *time.Time `json:"read_at"`

	// 创建时间，GORM 自动填
	CreatedAt time.Time `json:"created_at"`

	// 更新时间，GORM 自动维护
	UpdatedAt time.Time `json:"updated_at"`

	// 软删除：删除时不真删，只打 deleted_at 时间戳
	// json:"-" → 不返回给前端
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// ===== 下面是"假字段"（不存数据库），留给 M2/M3 拼装用 =====

	// 对端用户简介：M2 列表里要显示"对方是谁"
	// gorm:"-" → 不建列，由 service 层查出来填进去
	Peer *UserBrief `gorm:"-" json:"peer,omitempty"`
}
