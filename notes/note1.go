// 类比int main()
package notes

//类比#include<bits/stdc++.h>
import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// func main为程序起点
func main() {

	//创建了一个路由，叫r(router,路由)
	r := gin.Default()

	// 1. c.Query 获取 URL 参数  /search?keyword=校园卡

	//用get请求传参；func(c *gin.Context)：这个匿名函数就是服务员
	r.GET("/search", func(c *gin.Context) {

		//c.Query("keyword")：管家翻看记事本，去寻找 URL 里 ? 后面的参数。并将该参数赋值给keyword
		keyword := c.Query("keyword") // 如果没传，返回 ""

		//c.JSON:这里指定了返回的数据格式是 JSON（前后端对接的标准格式）。
		//gin.H{...}：这是 Gin 提供的一个偷懒语法糖。它本质上是 map[string]interface{}。
		// 它把你传入的键值对 ("type": "Query" 和 "value": keyword) 序列化成 JSON 格式。
		c.JSON(http.StatusOK, gin.H{"type": "Query", "value": keyword})
	})

	// 2. c.Param 获取路径参数  /item/123

	//:id 是一个占位符,相当于餐厅里“3号桌”、“5号桌”的桌子编号.顾客的访问路径可能是 /item/123，也可能是 /item/abc。
	r.GET("/item/:id", func(c *gin.Context) {

		//服务员（c）翻开记事本，提取出冒号后面那个真实的值。如果顾客访问的是 /item/123，那么 id 就是 "123"(string类型)。
		id := c.Param("id")

		//返回一个 JSON 格式的数据：{"type": "Param", "value": "123"}。
		c.JSON(http.StatusOK, gin.H{"type": "Param", "value": id})
	})

	// 3. c.PostForm 获取表单数据（后面学上传图片再用，现在知道有就行）

	//POST 意味着顾客在“提交”数据，而不是简单的“获取”数据。这里规定，顾客必须向 /form 这个地址提交数据。
	r.POST("/form", func(c *gin.Context) {

		//顾客在“表格”里填了一个叫 name 的选项。服务员打开顾客交上来的表格，寻找 name 这一栏，把里面的值提取出来。
		name := c.PostForm("name")

		//把提取到的名字返回给顾客。
		c.JSON(http.StatusOK, gin.H{"type": "Form", "value": name})
	})

	//保持程勋运行，直到按下Ctrl+C停止
	r.Run(":8000")
}
