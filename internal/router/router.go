// Package router 集中注册所有 HTTP 路由。
//
// ⚠️ 当前状态：这个包还没有被 main.go 调用。
// 真正生效的路由注册在 main.go 里，本文件里的路由【暂时不生效】。
//
// 所以：新加路由请先加在 main.go 里。
//
// 将来要切过来的时候，三件事一起做：
//  1. main.go 里改成 router.RegisterRoutes(r)
//  2. 删掉 main.go 里内联的那三个失物接口（lost-item / lost-items / lost-items/:id）
//     —— 它们已经搬到 internal/handler/lost_item.go 了，不删会重复注册，gin 会 panic
//  3. 删掉 main.go 里的 register / login / me 三条（本文件里已经有了）
//
// 切换完成后可以删掉这段注释。
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
