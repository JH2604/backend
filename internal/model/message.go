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

// MarkReadReq：M5 标记已读的请求体（PUT /messages/read）
// 三个字段是"三种范围"，优先级 ids > peer_id > all，由 service 层按顺序判断
type MarkReadReq struct {
	// 要标记的消息 ID 列表；不传就是空切片，落到下面两个范围
	// max=500：一次最多 500 条，防止前端误传超大数组拖慢 SQL
	IDs []uint `json:"ids" binding:"omitempty,max=500"`

	// 对端用户 ID：把"和这个人的会话"里的未读全部标记为已读
	// 用指针是为了区分"没传"(nil) 和"传了 0"：
	// 普通 uint 拿不到 null，缺字段和值 0 在 Go 里长得一模一样
	PeerID *uint `json:"peer_id"`

	// 是否清空当前用户的全部未读
	// bool 零值天生是 false（= 不是全部），所以不需要指针
	All bool `json:"all"`
}
