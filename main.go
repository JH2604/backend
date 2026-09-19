package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gin-demo/internal/model" // 引入新建的 model 包
)

// 2. 全局数据库连接对象
var db *gorm.DB

// 3. 初始化数据库连接
func initDB() {
	// 注意：你的密码现在改成了 123456，这里同步更新
	dsn := "root:123456@tcp(127.0.0.1:3306)/lost_found_db?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("❌ 数据库连接失败，错误信息: " + err.Error())
	}

	// 连接池配置，防止 invalid connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ⚠️ 关键：这里要改成 model.LostItem
	err = db.AutoMigrate(&model.LostItem{})
	if err != nil {
		panic("❌ 自动建表失败")
	}
	fmt.Println("✅ 数据库连接成功，表结构已同步！")
}

func main() {
	// 4. 启动时先连数据库
	initDB()

	r := gin.Default()

	// 5. 发布帖子的接口
	r.POST("/api/v1/lost-item", func(c *gin.Context) {
		var item model.LostItem // ⚠️ 这里改成 model.LostItem

		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "参数错误: " + err.Error(),
			})
			return
		}

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
			"data":    item,
		})
	})

	// 6. 获取失物招领列表
	r.GET("/api/v1/lost-items", func(c *gin.Context) {
		var items []model.LostItem // ⚠️ 这里改成 []model.LostItem

		location := c.Query("location")
		query := db.Model(&model.LostItem{}) // ⚠️ 这里改成 &model.LostItem{}

		if location != "" {
			query = query.Where("location LIKE ?", "%"+location+"%")
		}

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
		id := c.Param("id")
		var item model.LostItem // ⚠️ 这里改成 model.LostItem

		if err := db.First(&item, id).Error; err != nil {
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
