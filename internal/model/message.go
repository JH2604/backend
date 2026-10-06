package model

import (
	"time"

	"gorm.io/gorm"
)

// Message 私信实体（表 messages）
type Message struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SenderID   uint           `gorm:"index;not null" json:"sender_id"`
	ReceiverID uint           `gorm:"index:idx_receiver_read;not null" json:"receiver_id"`
	PostID     *uint          `gorm:"index" json:"post_id"` // 无关联帖子存 NULL，所以用指针
	Content    string         `gorm:"type:text;not null" json:"content"`
	IsRead     bool           `gorm:"index:idx_receiver_read;not null;default:false" json:"is_read"`
	ReadAt     *time.Time     `json:"read_at"`
	Reminded   bool           `gorm:"not null;default:false" json:"reminded"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Peer *UserBrief `gorm:"-" json:"peer,omitempty"` // 假字段，由 service 拼装
}

// MarkReadReq M5 标记已读请求体，三种范围优先级 ids > peer_id > all
type MarkReadReq struct {
	IDs    []uint `json:"ids" binding:"omitempty,max=500"`
	PeerID *uint  `json:"peer_id"` // 指针区分"未传"与 0
	All    bool   `json:"all"`
}

// box 参数取值（文档原文是 received，不是 inbox）
const (
	MessageBoxAll      = "all"
	MessageBoxSent     = "sent"
	MessageBoxReceived = "received"
)

// 每行消息相对"我"的方向，也是返回给前端的 direction
const (
	MessageDirectionSent     = "sent"
	MessageDirectionReceived = "received"
)

// MessageListQuery M2 的 query 参数
type MessageListQuery struct {
	Box string `form:"box" binding:"omitempty,oneof=all sent received"`
	// IsRead 用指针：区分"未传"（nil，返回全部）与显式 is_read=false
	IsRead   *bool `form:"is_read"`
	Page     int   `form:"page" binding:"omitempty,gte=1"`
	PageSize int   `form:"page_size" binding:"omitempty,gte=1,lte=50"`
}

// MessagePostBrief M2 列表项里的 post 字段（外层用 *MessagePostBrief，可为 null）
type MessagePostBrief struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// MessageListItem M2 列表项，字段严格按文档 schema 来，不直接返回实体
type MessageListItem struct {
	ID        uint              `json:"id"`
	Direction string            `json:"direction"`
	Peer      *UserBrief        `json:"peer"` // 我发出的 → 收件人；我收到的 → 发件人
	Content   string            `json:"content"`
	Post      *MessagePostBrief `json:"post"`
	// IsRead 随方向变化：我收到的 = 我是否已读；我发出的 = 对方是否已读
	IsRead    bool      `json:"is_read"`
	Reminded  bool      `json:"reminded"`
	CreatedAt time.Time `json:"created_at"`
}

// SendMessageReq M3 请求体（发送者由 token 确定）
type SendMessageReq struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required,notblank,max=1000"`
	PostID     *uint  `json:"post_id"` // 指针区分"未传"与 0
	Remind     bool   `json:"remind"`
}

// MessageView M3 的 data.message 与 M4 的 list 项共用
type MessageView struct {
	ID        uint      `json:"id"`
	Direction string    `json:"direction"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	Reminded  bool      `json:"reminded"`
	CreatedAt time.Time `json:"created_at"`
}

// 提醒结果枚举
const (
	RemindStatusSent    = "sent"
	RemindStatusSkipped = "skipped"
	RemindStatusFailed  = "failed"

	RemindChannelSMS   = "sms"
	RemindChannelEmail = "email"

	RemindReasonNotRequested = "not_requested"
	RemindReasonNoContact    = "no_contact"
	RemindReasonDisabled     = "disabled"
	RemindReasonRateLimited  = "rate_limited"
	RemindReasonProviderErr  = "provider_error"
)

// RemindResult M3 的 data.remind（sent 时 reason 为空，非 sent 时 channel 为空）
type RemindResult struct {
	Status  string `json:"status"`
	Channel string `json:"channel"`
	Reason  string `json:"reason"`
}

// SendMessageData M3 的 data
type SendMessageData struct {
	Message MessageView  `json:"message"`
	Remind  RemindResult `json:"remind"`
}

// ConversationQuery M4 的 query 参数（游标分页，不用 page/page_size）
type ConversationQuery struct {
	BeforeID *uint `form:"before_id"`
	Limit    int   `form:"limit" binding:"omitempty,gte=1,lte=50"`
	MarkRead bool  `form:"mark_read"`
}

// ConversationResult M4 的 data
type ConversationResult struct {
	Peer      UserBrief     `json:"peer"`
	CanRemind bool          `json:"can_remind"` // 规则同 U7，共用 service.canRemind()
	List      []MessageView `json:"list"`       // 时间正序（与 M2 的倒序相反）
	HasMore   bool          `json:"has_more"`
}
