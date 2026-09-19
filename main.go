package main

import (
	"fmt"
	"net/http"
	"time" // ⚠️ 新增这个

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 1. 数据库模型：多了 ID 字段
type LostItem struct {
	//uint为无符号整数，即没有负数的int，且数据范围是int的2倍
	//gorm:"primaryKey":告诉gorm，id是主键，类比数组的每个元素唯一下标
	ID uint `gorm:"primaryKey" json:"id"` // GORM 默认认为 ID 是主键，自增

	//binding:"required":它告诉 Gin 框架：“当前端发来 JSON 时，这个字段绝对不能为空，如果没有，直接报错拦截！”
	Title    string `json:"title" binding:"required"`
	Location string `json:"location"`
	Desc     string `json:"desc"`
}

// 2. 全局数据库连接对象
// 声明变量db，全程database(数据库),gorm.DB 是 GORM 库里的一个结构体,它里面装满了连接池、SQL 构建器、配置项等数据。
// *表示为指针，避免重复拷贝
var db *gorm.DB

// 3. 初始化数据库连接
func initDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/lost_found_db?charset=utf8mb4&parseTime=True&loc=Local"

	var err error

	//调用函数 gorm.Open(...);mysql.Open(dsn)：创建 MySQL 的“连接方言”，告诉 GORM 我们要连的是 MySQL，DSN 是地址。&gorm.Config{}：GORM 的配置对象。
	//函数会尝试用 TCP 协议去连接你本地的 MySQL 3306 端口，验证账号密码
	//函数有两个返回对象，一个指向 gorm.DB 结构体的指针(结果对象),一个error 接口的值(错误对象)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		//panic：相当于 C++ 里的 std::abort() 或者抛出无法捕获的致命异常。程序直接打印红色错误信息并退出，后面的 r.Run(":8000") 连启动的机会都没有。
		//err.Error()：调用 error 接口里的方法，把具体的错误信息转成字符串
		panic("❌ 数据库连接失败，错误信息: " + err.Error())
	}

	// ⚠️ 新增：获取底层 sql.DB 连接池，防止 invalid connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)           // 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxOpenConns(100)          // 设置打开数据库连接的最大数量
		sqlDB.SetConnMaxLifetime(time.Hour) // 设置了连接可复用的最大时间（1小时）
	}

	err = db.AutoMigrate(&LostItem{})
	if err != nil {
		panic("❌ 自动建表失败")
	}
	fmt.Println("✅ 数据库连接成功，表结构已同步！")
}

func main() {
	// 4. 启动时先连数据库
	initDB()

	r := gin.Default()

	r.GET("/api/items/:id", handler.GetItem)
	r.GET("/api/items", handler.ListItems)

		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "参数错误: " + err.Error(),
			})
			return
		}

		// 核心魔法：将 item 写入数据库
		// db.Create(&item) 相当于 SQL 的 INSERT INTO lost_items (...) VALUES (...)
		if err := db.Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "数据库写入失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "发布成功，已存入数据库",
			"data":    item, // 此时 item 里多了一个数据库生成的 ID
		})
	})

	// 6. 获取失物招领列表
	r.GET("/api/v1/lost-items", func(c *gin.Context) {
		var items []LostItem

		// 获取 URL 参数，例如 /api/v1/lost-items?location=图书馆
		location := c.Query("location")

		// 构建查询：db 是 GORM 的入口
		query := db.Model(&LostItem{})

		// 如果前端传了 location 参数，就加上筛选条件
		if location != "" {
			query = query.Where("location LIKE ?", "%"+location+"%") // LIKE 模糊匹配
		}

		// 执行查询（相当于 SELECT * FROM lost_items）
		if err := query.Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "查询成功",
			"data":    items,
		})
	})

	// 7. 获取单条帖子详情
	r.GET("/api/v1/lost-items/:id", func(c *gin.Context) {
		id := c.Param("id") // 获取路径里的 id
		var item LostItem

		// 根据主键查询（相当于 SELECT * FROM lost_items WHERE id = ?）
		if err := db.First(&item, id).Error; err != nil {
			// ⚠️ 极其重要：判断是不是“没找到”的错误
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "帖子不存在",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "查询失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "查询成功",
			"data":    item,
		})
	})
	r.Run(":8000")
}
