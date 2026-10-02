package router

import (
	"gin-demo/internal/handler"
	"gin-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 负责注册所有的 HTTP 路由
func RegisterRoutes(r *gin.Engine) {
	// 队友负责的用户模块路由
	r.POST("/api/v1/auth/register", handler.Register) //注册
	r.POST("/api/v1/auth/login", handler.Login)
	r.GET("/api/v1/users/me", middleware.Auth(), handler.GetCurrentUser)
	r.POST("/api/v1/auth/logout",middleware.Auth(),handler.Logout)
	r.POST("/api/v1/files", middleware.Auth(), handler.UploadPhoto)
	r.POST("/api/v1/auth/refresh", handler.Refresh)

	// ---------- 你负责的失物招领模块 ----------
	v1 := r.Group("/api/v1")
	{
		v1.POST("/posts", handler.CreateLostItem)
		v1.GET("/posts", handler.ListLostItems)
		v1.GET("/posts/:id", handler.GetLostItem)
	}
}
