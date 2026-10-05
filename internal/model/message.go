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

	// 来源帖子 ID → posts.id。可以为空（数据库里是 NULL）
	// 从"帖子详情页"点【联系TA】发起的私信才有值；别处发起的没有关联帖子
	// 为什么用指针 *uint：没有关联帖子时必须存 NULL，
	// 用普通 uint 存 0 会被误读成"关联了 id=0 的帖子"，前端也会拿到 0 而不是 null
	PostID *uint `gorm:"index" json:"post_id"`

	// 私信内容
	Content string `gorm:"type:text;not null" json:"content"`

	// 是否已读。默认 false（刚发出来就是未读）
	// 和 ReceiverID 共用复合索引，服务于 M1 的"未读计数"
	IsRead bool `gorm:"index:idx_receiver_read;not null;default:false" json:"is_read"`

	// 已读时间。未读时为 nil，序列化成 null
	ReadAt *time.Time `json:"read_at"`

	// 这条私信发送时是否触发过短信 / 邮件提醒
	// M2 只是把它读出来返回；真正写 true 的是 M3（POST /messages 传 remind=true）
	Reminded bool `gorm:"not null;default:false" json:"reminded"`

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

// ============ M2 我的消息列表（GET /messages）============

// 收件箱范围：query 参数 box 的三个取值（文档原文，received 不是 inbox）
const (
	MessageBoxAll      = "all"      // 发出的 + 收到的（默认）
	MessageBoxSent     = "sent"     // 只看我发出的
	MessageBoxReceived = "received" // 只看我收到的
)

// 每行消息相对【我】的方向，也是返回给前端的 direction 字段
// 和上面 box 的值域长得一样，但含义不同：
// box 是"筛选条件"，direction 是"这一行是发还是收"，所以分开定义
const (
	MessageDirectionSent     = "sent"
	MessageDirectionReceived = "received"
)

// MessageListQuery：M2 的 query 参数
// GET 用 form 标签 + c.ShouldBindQuery（和 P1 的 ListPostsQuery 一个套路）
type MessageListQuery struct {
	// Box：不传就是"全部"。oneof 只放文档给的三个值，传 inbox 直接 40000
	Box string `form:"box" binding:"omitempty,oneof=all sent received"`

	// IsRead：为什么是【指针 *bool】而不是 bool？
	// 文档要求"不传返回全部"，也就是"只看未读"和"只看已读"之外还有第三种状态。
	// 普通 bool 的零值是 false，"用户没传"和"用户显式传了 is_read=false"在 Go 里长得一模一样，
	// 分不出来就会把"不传"错当成"只看已读"。指针有第三个值 nil 表示"没传"
	IsRead *bool `form:"is_read"`

	// Page / PageSize：复用 P1 已经定好的默认值和上限（1 / 20 / 50）
	Page     int `form:"page" binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=50"`
}

// MessagePostBrief：列表项里的 post 字段
// 从帖子详情页发起的私信才不为 null，所以外层用 *MessagePostBrief
type MessagePostBrief struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// MessageListItem：M2 列表里的一项，字段严格按文档 schema Message 来，不多不少
// 为什么不直接返回 model.Message？
// ① 会多吐出 sender_id / receiver_id / read_at，文档里没有这些字段；
// ② 文档要的是"对端 peer + 方向 direction"，前端不该自己算"对面是发件人还是收件人"；
// ③ 和 P1 的 PostListItem 保持同一种做法：实体只进数据库，出网关注走 DTO
type MessageListItem struct {
	ID uint `json:"id"`

	// Direction：这一行相对"我"是发出还是收到（MessageDirectionSent / Received）
	Direction string `json:"direction"`

	// Peer：对端是谁。我发出的 → 收件人；我收到的 → 发件人
	Peer *UserBrief `json:"peer"`

	Content string `json:"content"`

	// Post：关联帖子，没有就是 null
	Post *MessagePostBrief `json:"post"`

	// IsRead：直接映射 messages.is_read 这一列。
	// 它的含义随方向自动变化：我收到的消息 → 我是否已读；我发出的消息 → 对方是否已读
	IsRead bool `json:"is_read"`

	// Reminded：是否触发过短信 / 邮件提醒
	Reminded bool `json:"reminded"`

	CreatedAt time.Time `json:"created_at"`
}
