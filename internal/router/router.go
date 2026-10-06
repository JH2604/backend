package router

import (
	"gin-demo/internal/handler"
	"gin-demo/internal/middleware"
	"gin-demo/pkg/validate"
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
	// ========== 用户模块 ==========
	r.POST("/api/v1/auth/register", handler.Register)                         // A1 注册（学号去实名库换真名）
	r.POST("/api/v1/auth/login", handler.Login)                               // A2 登录：发 access + refresh 双令牌
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)      // U1 当前登录用户的完整资料 + 帖子数
	r.GET("/api/v1/users/admins", middleware.Auth(), handler.ListAdmins)      // U6 联系管理员：管理员列表
	r.GET("/api/v1/users/:user_id", middleware.Auth(), handler.GetUserPublic) // U7 查看发帖人信息（手机号等 detail 只给管理员）
	r.POST("/api/v1/auth/logout", middleware.Auth(), handler.Logout)          // A4 退出登录（吊销当前会话）
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)           // F1 上传图片
	r.POST("/api/v1/auth/refresh", handler.Refresh)                           // A3 刷新令牌（旧的刷新令牌一次性，用过就换新）

	// ========== 失物招领模块 ==========
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
		v1.POST("/messages", middleware.Auth(), handler.SendMessage)                // M3 发送私信（可选提醒）
		// M4 路径不能写成 /messages/:peer_id —— 会和 M1 的 /messages/unread-count 冲突，gin 启动即 panic
		v1.GET("/messages/conversations/:peer_id", middleware.Auth(), handler.GetConversation) // M4 与某用户的私信记录（游标分页）

		// ---- 改密码 ----
		v1.PUT("/users/me/password", middleware.Auth(), handler.ChangePassword) // U3 改密码（成功后吊销全部会话，要重新登录）
		v1.PATCH("/users/me", middleware.Auth(), handler.UpdateAvatar)
	}
}

// 把字段值去掉两端空白，还剩下东西就算合格
func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", func(fl validator.FieldLevel) bool {
			return strings.TrimSpace(fl.Field().String()) != ""
		})
		v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			return validate.Password(fl.Field().String())
		})
		v.RegisterValidation("studentid", func(fl validator.FieldLevel) bool {
			return validate.StudentID(fl.Field().String())
		})
	}
}
