package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"campus-lost-found/backend/errcode"
	"campus-lost-found/backend/middleware"
	"campus-lost-found/backend/models"
	"campus-lost-found/backend/response"
)

func main() {
	// gin.New() = 空的路由器（不带任何中间件）
	r := gin.New()

	// 手动挂上中间件 —— 顺序很重要，从上到下依次执行
	r.Use(middleware.Logger())   // ① 先记日志
	r.Use(middleware.Recovery()) // ② 再捕获异常
	r.Use(middleware.Counter())  // ③ 统计请求数
	r.Use(middleware.OnlyGet())

	r.GET("/api/items/1", func(c *gin.Context) {
		item := models.Item{ID: 1, Name: "黑色雨伞", Type: "lost", Status: "pending"}
		response.Success(c, item)
	})

	r.GET("/api/items/999", func(c *gin.Context) {
		response.Fail(c, errcode.ErrNotFound)
	})

	// 故意写一个会崩溃的接口，用来测试
	r.GET("/api/crash", func(c *gin.Context) {
		var arr []int
		fmt.Println(arr[10]) // 越界 → panic
	})

	r.Run(":8080")
}
