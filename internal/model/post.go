package model

import (
	"time"

	"gorm.io/gorm"
)

// Post：帖子实体，对应数据库表 posts
// 设计依据：① 前端原型字段 ② 队友确认的契约

type Post struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// 作者 ID → users.id。由服务端从 token 取，前端不许传
	// index：以后"某用户发的帖子"（P6 / mine=true）会走索引
	UserID uint `gorm:"index;not null" json:"user_id"`

	// 类型：lost(失物) / found(拾物)。
	// size 限制长度，确保索引不会太长，避免查找时爆炸
	Type string `gorm:"size:10;index;not null" json:"type"`

	// 标题
	Title string `gorm:"size:100;not null" json:"title"`

	// 正文。text → MySQL TEXT，适合长文本
	//type区分lost/found种类，加快检索速度
	Content string `gorm:"type:text" json:"content"`

	// 图片：存【相对路径】数组，如 ["/uploads/xxx.png"]（前端自己拼域名）
	// serializer:json → GORM 自动把这个 []string 序列化成 JSON 存进一列，读时自动还原
	Images []string `gorm:"serializer:json;type:json" json:"images"`

	// 地点对象。 字段名严格按文档：latitude / longitude（不是 lat / lng）
	Location PostLocation `gorm:"serializer:json;type:json" json:"location"`

	// 事件发生时间。前端传 RFC3339，如 2026-09-22T15:04:05+08:00
	EventTime *time.Time `gorm:"index" json:"event_time"`

	// 状态：只有两个值  open(待处理) / closed(已完结)
	//default:如果前端没传status,数据库自动把它设为open
	Status string `gorm:"size:10;index;not null;default:'open'" json:"status"`

	//插入和修改时间时自动推进时间
	CreatedAt time.Time `json:"created_at"` // GORM 自动填
	UpdatedAt time.Time `json:"updated_at"` // GORM 自动维护

	// 标记为已找到 / 已认领的时间。进行中为 nil，序列化成 null
	ClosedAt *time.Time `json:"closed_at"`

	//  软删除：删除时不真的 DELETE，而是把 deleted_at 打上时间戳
	//    GORM 之后所有查询都会自动加上 WHERE deleted_at IS NULL（只查没被删的）
	//    （所以"已删除的帖子"不会出现在列表/详情里，你一行代码都不用加）
	// json:"-" → 这个字段不返回给前端
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 管理员删帖的原因（软删除配套留痕）。json:"-" 不暴露给普通接口
	//json:"-"：这句话是精髓！ 它告诉 Gin 框架：“这个字段是后端的内部秘密，千万不要返回给前端！”
	DeleteReason string `gorm:"size:200" json:"-"`

	//这两个字段是“假字段”。它们不存进数据库，
	// 只是为了前端展示时，能够顺便把“作者信息”和“评论总数”拼装在一起返回给前端，省得前端来回请求。

	// gorm:"-" 告诉 GORM：别为我建列、别在 SQL 里管我
	// 由 service 层批量查出来填进去（避免 N+1，见下方说明）
	Author *UserBrief `gorm:"-" json:"author,omitempty"`

	IsMine          bool `gorm:"-" json:"is_mine"`
	CanDelete       bool `gorm:"-" json:"can_delete"`
	CanChangeStatus bool `gorm:"-" json:"can_change_status"`
}

// PostLocation 地点对象（JSON 列）
type PostLocation struct {
	Name string `json:"name"`

	//这个 omitempty 是告诉 Gin：“如果前端没传这个值，
	// 或者它等于零值（0），就不要把它序列化到 JSON 里返回给我。”
	//Latitude（纬度）
	Latitude float64 `json:"latitude,omitempty"`
	//Longitude（经度)
	Longitude float64 `json:"longitude,omitempty"`
}

// UserBrief 用户"简介"结构：列表里展示作者用
// ⚠️ 绝不能直接返回 model.User —— 那里有 Password 字段，会泄露密码哈希！
type UserBrief struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`       // 对应 User.Username
	AvatarURL string `json:"avatar_url"` // 现在还没有头像字段，先留空

	//role对应不同权限
	Role string `json:"role"`
}

// ============ 枚举常量：集中定义，写错编译器会提醒 ============
const (
	PostTypeLost  = "lost"
	PostTypeFound = "found"
)

const (
	PostStatusOpen   = "open"   // 待处理
	PostStatusClosed = "closed" // 已完结
)

// 分页默认值（ 按队友：默认 20，最大 50）
/*
const ( ... )：这是 Go 语言声明常量的语法。
常量的意思是“一旦定义，后续代码里就不能修改它”。
（C++ 里的 const int DefaultPage = 1; 效果一样）。

DefaultPage = 1：当前端请求帖子列表，但没有传页码时，后端默认返回第 1 页。

DefaultPageSize = 20：当前端没有传“每页多少条”时，后端默认每页给 20 条帖子。
为什么是 20？因为 20 条数据刚好填满手机屏幕一两屏，加载快，流量也不大。

MaxPageSize = 50：这是极其重要的安全底线！ 假如坏人前端故意传一个 page_size = 100000，
如果后端不加限制，数据库就会一次性吐出十万条数据，直接导致内存溢出（OOM），
服务器崩溃。加上这个限制，后端最多只给 50 条。
*/
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 50
	DefaultOrder    = "desc"
)

// CreatePostReq：P3 发布帖子
// 注意没有 UserID —— 作者由服务端从 token 取，前端在物理上无法伪造
type CreatePostReq struct {
	Type      string       `json:"type" binding:"required,oneof=lost found"`
	Title     string       `json:"title" binding:"required,notblank,max=30"`
	Content   string       `json:"content" binding:"required,notblank,max=1000"`
	Images    []string     `json:"images" binding:"omitempty,max=9"`
	Location  PostLocation `json:"location" binding:"required"`
	EventTime *time.Time   `json:"event_time"` // RFC3339
}

// DeletePostReq：P4 删除（可选请求体）
type DeletePostReq struct {
	Reason string `json:"reason" binding:"omitempty,max=200"`
}

// UpdatePostStatusReq：P5 改状态（只有两个值）
type UpdatePostStatusReq struct {
	Status string `json:"status" binding:"required,oneof=open closed"`
}

// ListPostsQuery：P1 / P6 的查询参数
// ⭐ GET 用 form 标签 + c.ShouldBindQuery；POST 才用 json 标签 + ShouldBindJSON
/*
核心前提：为什么是 form 标签？
前端发起 GET 请求时，参数是拼在网址后面的，比如：
/api/v1/posts?page=1&page_size=20&type=lost&keyword=校园卡
这种格式叫 Query 参数。Gin 框架里，
接收这种参数必须用 c.ShouldBindQuery(&query)，
并且结构体字段必须打 form 标签（不是 json 标签）。
*/

type ListPostsQuery struct {
	//page:要查第几页
	Page int `form:"page" binding:"omitempty,gte=1"`
	//pagesize:一页显示多少条数据，最少1最多50
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=50"`
	//type:筛选帖子类型
	Type string `form:"type" binding:"omitempty,oneof=all lost found"`
	//keyword:按照哪个关键词筛选
	Keyword string `form:"keyword" binding:"omitempty,max=50"`
	//mine是否只看自己发的帖子
	Mine bool `form:"mine"`
	//status:open和close的状态筛选
	Status string `form:"status" binding:"omitempty,oneof=all open closed"`
	//order:升序/降序
	Order string `form:"order" binding:"omitempty,oneof=asc desc"`
}

type PostListItem struct {
	ID             uint         `json:"id"`
	Type           string       `json:"type"`
	Title          string       `json:"title"`
	ContentPreview string       `json:"content_preview"`
	CoverURL       *string      `json:"cover_url"`
	ImageCount     int          `json:"image_count"`
	Location       PostLocation `json:"location"`
	Status         string       `json:"status"`
	Author         *UserBrief   `json:"author,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	ClosedAt       *time.Time   `json:"closed_at"`
}

type PhotoSize struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int64  `json:"size"`
}
