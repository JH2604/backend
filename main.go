package main

import (
	"fmt"
	"time"

	"gin-demo/internal/middleware"
	"gin-demo/internal/repository"
	"gin-demo/internal/router"
	"gin-demo/pkg/config"

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
	err = db.AutoMigrate(&model.Post{}, &model.User{}, &model.Session{})
	if err != nil {
		panic("❌ 自动建表失败")
	}
	fmt.Println("✅ 数据库连接成功，表结构已同步！")
}

func main() {
	if config.JWTSecret == "" {
		panic("❌ 必须设置环境变量 JWT_SECRET(长度 ≥ 32 字节）")
	}
	// 4. 启动时先连数据库
	initDB()

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	router.RegisterRoutes(r)
	r.Static("/uploads", "./uploads")

	r.Run(":8000")
}
