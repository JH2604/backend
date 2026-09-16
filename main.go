package main

import (
	"fmt"

	"campus-lost-found/backend/handler"

	"github.com/gin-gonic/gin"

	"campus-lost-found/backend/middleware"
)

func main() {
	// gin.New() = 空的路由器（不带任何中间件）
	r := gin.New()

	// 手动挂上中间件 —— 顺序很重要，从上到下依次执行
	r.Use(middleware.Logger())   // ① 先记日志
	r.Use(middleware.Recovery()) // ② 再捕获异常
	r.Use(middleware.Counter())  // ③ 统计请求数
	r.Use(middleware.OnlyGet())

	r.GET("/api/items/:id", handler.GetItem)
	r.GET("/api/items", handler.ListItems)

	// 故意写一个会崩溃的接口，用来测试
	r.GET("/api/crash", func(c *gin.Context) {
		var arr []int
		fmt.Println(arr[10]) // 越界 → panic
	})

	r.Run(":8080")
}
