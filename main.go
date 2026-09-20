package main

import (
	"fmt"
	"time"

	"gin-demo/internal/handler"
	"gin-demo/internal/repository"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

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
	repository.Init(db)

	// 连接池配置，防止 invalid connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ⚠️ 关键：这里要改成 model.LostItem
	err = db.AutoMigrate(&model.LostItem{}, &model.User{})
	if err != nil {
		panic("❌ 自动建表失败")
	}
	fmt.Println("✅ 数据库连接成功，表结构已同步！")
}

func main() {
	// 4. 启动时先连数据库
	initDB()

	r := gin.Default()
	r.POST("/api/register", handler.Register)

	// 5. 发布帖子的接口
	r.POST("/api/v1/lost-item", func(c *gin.Context) {
		var item model.LostItem // ⚠️ 这里改成 model.LostItem

		if err := c.ShouldBindJSON(&item); err != nil {
			response.FailReason(c, errcode.ErrInvalidParams, err.Error())
			fmt.Println("❌ 发布失败:", err.Error())
			return
		}

		if err := db.Create(&item).Error; err != nil {
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 数据写入失败:", err.Error())
			return
		}

		response.Success(c, item)

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
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 失物查询失败:", err.Error())
			return
		}

		response.Success(c, items)
	})

	// 7. 获取单条帖子详情
	r.GET("/api/v1/lost-items/:id", func(c *gin.Context) {
		id := c.Param("id")
		var item model.LostItem // ⚠️ 这里改成 model.LostItem

		if err := db.First(&item, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Fail(c, errcode.ErrNotFound)
				return
			}
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 帖子查询失败:", err.Error())
			return
		}

		response.Success(c, item)
	})

	r.Run(":8000")
}
