package router

import (
	"gin-demo/internal/handler"
	"gin-demo/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 负责注册所有的 HTTP 路由
func RegisterRoutes(r *gin.Engine) {
	// 队友负责的用户模块路由
	r.POST("/api/register", handler.Register) //注册
	r.POST("/api/login", handler.Login)
	r.GET("/api/me", middleware.Auth(), handler.GetCurrentUser)

	// 你负责的失物招领模块路由
	v1 := r.Group("/api/v1")
	{
		v1.POST("/lost-items", handler.CreateLostItem)
		v1.GET("/lost-items", handler.ListLostItems)
		v1.GET("/lost-items/:id", handler.GetLostItem)
	}
}
