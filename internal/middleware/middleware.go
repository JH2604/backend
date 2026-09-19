package middleware

import (
	"fmt"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		fmt.Printf("收到请求: %s %s\n", method, path)

		c.Next()

		duration := time.Since(start)

		fmt.Printf("处理完成:%s %s| 耗时 %v\n", method, path, duration)

	}

}

// Recovery 捕获 panic，统一返回错误响应（作业要求）
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		// defer：不管 c.Next() 里发生什么，这段都会执行
		defer func() {
			// recover() 接住 panic，返回错误信息；没 panic 时返回 nil
			if err := recover(); err != nil {
				fmt.Println("❌ 捕获到异常:", err)
				debug.PrintStack() // 打印堆栈，方便排查

				response.Fail(c, errcode.ErrServer) // 统一返回错误响应
				c.Abort()                           // 终止后续处理
			}
		}()

		c.Next() // 执行后面的中间件和处理函数
	}
}

func Counter() gin.HandlerFunc {
	count := 0
	return func(c *gin.Context) {
		count++
		fmt.Printf("这是第%d个请求\n", count)
		c.Next()

	}
}

func OnlyGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			response.Fail(c, errcode.ErrForbidden)

			c.Abort()
			return

		}
		c.Next()

	}

}
