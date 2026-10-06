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
func RegisterRoutes(r *gin.Engine) {
<<<<<<< Updated upstream
	// 队友负责的用户模块路由
	r.POST("/api/v1/auth/register", handler.Register) //注册
	r.POST("/api/v1/auth/login", handler.Login)
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)
	r.POST("/api/v1/auth/logout", middleware.Auth(), handler.Logout)
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)
	r.POST("/api/v1/auth/refresh", handler.Refresh)

	// ---------- 你负责的失物招领模块 ----------
=======
	r.POST("/api/v1/auth/register", handler.Register)                         // A1 注册（学号去实名库换真名）
	r.POST("/api/v1/auth/login", handler.Login)                               // A2 登录：发 access + refresh 双令牌
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)      // U1 当前登录用户的完整资料 + 帖子数
	r.GET("/api/v1/users/admins", middleware.Auth(), handler.ListAdmins)      // U6 我做的 · 联系管理员：管理员列表
	r.GET("/api/v1/users/:user_id", middleware.Auth(), handler.GetUserPublic) // U7 我做的 · 查看发帖人信息（手机号等 detail 只给管理员）
	r.POST("/api/v1/auth/logout", middleware.Auth(), handler.Logout)          // A4 退出登录（吊销当前会话）
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)           // F1 上传图片
	r.POST("/api/v1/auth/refresh", handler.Refresh)                           // A3 刷新令牌（旧的刷新令牌一次性，用过就换新）

>>>>>>> Stashed changes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/post", middleware.Auth(), handler.CreatePost)
		v1.GET("/posts", middleware.Auth(), handler.ListPosts)
		v1.GET("/posts/:id", middleware.Auth(), handler.GetPost)
		v1.PATCH("/posts/:id/status", middleware.Auth(), handler.UpdatePostStatus)
		v1.DELETE("/posts/:id", middleware.Auth(), handler.DeletePost)
		v1.GET("/messages/unread-count", middleware.Auth(), handler.GetUnreadCount)
		v1.PUT("/users/me/password", middleware.Auth(), handler.ChangePassword)

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
