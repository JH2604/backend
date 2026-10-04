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
	// 队友负责的用户模块路由
	r.POST("/api/v1/auth/register", handler.Register) //注册
	r.POST("/api/v1/auth/login", handler.Login)
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)
	r.POST("/api/v1/auth/logout", middleware.Auth(), handler.Logout)
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)
	r.POST("/api/v1/auth/refresh", handler.Refresh)

	// ---------- 你负责的失物招领模块 ----------
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/post", middleware.Auth(), handler.CreatePost)
		v1.GET("/posts", middleware.Auth(), handler.ListPosts)
		v1.GET("/posts/:id", middleware.Auth(), handler.GetPost)
		v1.PATCH("/posts/:id/status", middleware.Auth(), handler.UpdatePostStatus)
		v1.DELETE("/posts/:id", middleware.Auth(), handler.DeletePost)
		v1.GET("/messages/unread-count", middleware.Auth(), handler.GetUnreadCount)
		v1.PUT("/messages/read", middleware.Auth(), handler.MarkRead)
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
