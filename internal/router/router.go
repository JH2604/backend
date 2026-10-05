package router

import (
	"gin-demo/internal/handler"
	"gin-demo/internal/middleware"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterRoutes 负责注册所有的 HTTP 路由
// 关于 middleware.Auth()：就是"进门先查学生证"——验证请求头里的 Bearer token，
// 把"我是谁"塞进 context；后面 handler 里 middleware.GetUserID(c) 就是把它取出来。
// 登录 / 注册 / 刷新令牌这三条【不能】加 —— 那时候用户还没登录，没有 token 可查。
func RegisterRoutes(r *gin.Engine) {
	// ========== 用户模块（队友「风尽起长歌」负责：A 系列 + U1 + F1）==========
	// 这一块按队友原来的顺序排，整块归他；中间只有标了 ★ 的 U6/U7 是我写的，
	// 验收时按这个分人问：没标的是队友的，★ 是我的
	r.POST("/api/v1/auth/register", handler.Register)                         // A1 注册（学号去实名库换真名）
	r.POST("/api/v1/auth/login", handler.Login)                               // A2 登录：发 access + refresh 双令牌
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)      // U1 当前登录用户的完整资料 + 帖子数
	r.GET("/api/v1/users/admins", middleware.Auth(), handler.ListAdmins)      // ★ U6 我做的 · 联系管理员：管理员列表
	r.GET("/api/v1/users/:user_id", middleware.Auth(), handler.GetUserPublic) // ★ U7 我做的 · 查看发帖人信息（手机号等 detail 只给管理员）
	r.POST("/api/v1/auth/logout", middleware.Auth(), handler.Logout)          // A4 退出登录（吊销当前会话）
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)           // F1 上传图片
	r.POST("/api/v1/auth/refresh", handler.Refresh)                           // A3 刷新令牌（旧的刷新令牌一次性，用过就换新）

	// ========== 失物招领模块（我负责：帖子 P / 私信 M / 改密码 U3）==========
	// v1 是"路由组"：统一带上 /api/v1 前缀，所以组内路径不用再写一遍
	v1 := r.Group("/api/v1")
	{
		// ---- 帖子 ----
		v1.POST("/auth/post", middleware.Auth(), handler.CreatePost)               // 发帖（类型：失物/招领，可带图片）
		v1.GET("/posts", middleware.Auth(), handler.ListPosts)                     // P1 帖子列表：按类型/状态/关键词筛选 + 分页
		v1.GET("/posts/:id", middleware.Auth(), handler.GetPost)                   // 帖子详情：完整内容、图片、发帖人
		v1.PATCH("/posts/:id/status", middleware.Auth(), handler.UpdatePostStatus) // 改状态：进行中 / 已找到 / 已认领（仅楼主或管理员）
		v1.DELETE("/posts/:id", middleware.Auth(), handler.DeletePost)             // 删除帖子：软删 + 记录理由（仅楼主或管理员）

		// ---- 私信 ----
		v1.GET("/messages", middleware.Auth(), handler.ListMessages)                // M2 我的消息列表：方向/对端/关联帖子 + 筛选分页
		v1.GET("/messages/unread-count", middleware.Auth(), handler.GetUnreadCount) // M1 未读总数（前端红点用）
		v1.PUT("/messages/read", middleware.Auth(), handler.MarkRead)               // M5 标记已读：按 ids / peer_id / all 三种范围
		v1.POST("/messages", middleware.Auth(), handler.SendMessage)                // M3 发送私信：可选短信/邮件提醒（12h/人 1 次、每人每天 5 次）
		// ★ M4 路径必须是 /messages/conversations/:peer_id，不能写成 /messages/:peer_id ——
		// 后者会和上面 M1 的 /messages/unread-count 抢同一层通配符，gin 启动直接 panic
		v1.GET("/messages/conversations/:peer_id", middleware.Auth(), handler.GetConversation) // M4 与某用户的私信记录：游标分页、时间正序

		// ---- 改密码：功能属用户模块，但接口是我在这个组里实现的 ----
		v1.PUT("/users/me/password", middleware.Auth(), handler.ChangePassword) // U3 改密码（成功后吊销全部会话，要重新登录）
	}
}

// 把字段值去掉两端空白，还剩下东西就算合格
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})
	}
}
