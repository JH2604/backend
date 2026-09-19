package notes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ==========================================
// 1. 结构体定义：用来接收前端传来的 JSON 数据
// ==========================================
type LostItem struct {
	//首字母大写代表“公开”（Public），这样 Gin 框架才能在其他包里访问并修改它。如果小写，Gin 是看不见的。
	//json:"title"（反引号里的内容）：这是标签（Tag）。前端 Vue 发来的 JSON 是小写的 "title"，但 Go 的字段是大写的 Title
	Title    string `json:"title" binding:"required"` // binding:"required" 表示必填，不填会报错
	Location string `json:"location"`
	Desc     string `json:"desc"`
}

func main() {
	// 2. 初始化 Gin 引擎
	r := gin.Default()

	// ==========================================
	// 【阶段一】接收 URL 参数：Query 和 Param
	// ==========================================

	// 2.1 c.Query 获取 URL 参数  /search?keyword=手机
	r.GET("/search", func(c *gin.Context) {
		keyword := c.Query("keyword")

		//gin.H{...} —— 构造返回给前端的具体数据
		/*gin.H 是什么？ 你可以把它看作是一个语法糖。它本质上就是 map[string]interface{}（一个键是字符串，值可以是任何类型的 Map）。
		类似于 C++ 里的 std::unordered_map<std::string, std::any>。
		为什么要用它？因为它极其方便。你不需要专门去定义一个结构体（比如 ErrorResponse），
		直接 gin.H{"code": 400, "message": "..."} 就能快速拼装出一个 JSON 对象*/
		c.JSON(http.StatusOK, gin.H{
			"type":  "Query",
			"value": keyword,
		})
	})

	// 2.2 c.Param 获取路径参数  /item/123
	r.GET("/item/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"type":  "Param",
			"value": id,
		})
	})

	// ==========================================
	// 【阶段二】接收 POST 请求数据：Form 和 JSON
	// ==========================================

	// 3.1 c.PostForm 获取表单数据（传统网页表单）
	r.POST("/form", func(c *gin.Context) {
		name := c.PostForm("name")
		c.JSON(http.StatusOK, gin.H{
			"type":  "Form",
			"value": name,
		})
	})

	// 3.2 c.ShouldBindJSON 接收 JSON 数据（前后端分离的主流方式）
	//前端通过post的方式向"/api/v1/lost-item"传参,func为匿名函数，类比sort中的cmp，后续程序如果复杂需专门写一个函数
	r.POST("/api/v1/lost-item", func(c *gin.Context) {
		var item LostItem // 准备一个空的 LostItem 变量

		// 核心：把前端传来的 JSON 解析到 item 变量里
		//如果没有错误，err即为空(nil),有错的话nil就会显示错在哪里
		if err := c.ShouldBindJSON(&item); err != nil {
			// 如果解析失败（比如前端没传 title，或者 JSON 格式不对）
			//http.StatusBadRequest —— 设置 HTTP 状态码（400）
			/*200 代表一切顺利。400 代表你前端传的参数有问题，别甩锅给后端。500 代表后端代码自己出 Bug 了（比如空指针、数据库连不上）。*/
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "参数错误: " + err.Error(),
			})
			return //  极其重要：报错必须 return，不然代码会继续往下跑导致程序崩溃
		}

		// 解析成功，返回数据
		/*
			状态码常量	                     数值	  含义	               使用场景
			http.StatusOK	                200	     成功	            数据查询成功、注册成功、发帖成功
			http.StatusBadRequest	        400	     客户端参数错误	     前端少传了字段、JSON 格式不对
			http.StatusUnauthorized     	401	     未授权	             没登录就想删帖
			http.StatusNotFound	            404	     资源不存在	         访问了不存在的接口
			http.StatusInternalServerError	500	     服务器内部错误	     数据库崩了、代码空指针了
		*/
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "发布成功",
			"data":    item,
		})
	})

	// 4. 启动服务，监听 8000 端口
	r.Run(":8000")
}
