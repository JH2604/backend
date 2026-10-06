package main

import (
	"fmt"
	"time"

	//gin-demo为模块名,gin-demo 的根目录 = /Users/lovyy/Documents/backend/
	"gin-demo/internal/middleware"
	"gin-demo/internal/repository"
	"gin-demo/internal/router"
	"gin-demo/pkg/config"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql" // ← 内层：mysql.Open 来自这里（方言/驱动）
	"gorm.io/gorm"         // ← 外层：gorm.Open 来自这里（通用引擎）

	"gin-demo/internal/model" // 引入新建的 model 包
)

// 2. 全局数据库连接对象
var db *gorm.DB //声明一个指向gorm.db的空指针

// 3. 初始化数据库连接
func initDB() {
	// 注意：你的密码现在改成了 123456，这里同步更新
	//写好地址+账号+库名
	//[用户名[:密码]@][协议[(地址)]]/库名[?参数1=值1&参数2=值2...]
	/*
			   root    :   123456   @   tcp(127.0.0.1:3306)   /   lost_found_db   ?   charset=utf8mb4&parseTime=True&loc=Local
		   └─①─┘   └───②───┘   └───────③──────────┘   └───────④───────┘   └────────────⑤────────────┘
		    用户名    密码        分隔符      连接方式+地址:端口          数据库名              连接参数

	*/
	dsn := config.DatabaseDSN

	var err error
	//连接数据库
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("❌ 数据库连接失败，错误信息: " + err.Error())
	}
	//把连接好的对象交给repository，“依赖注入”
	repository.Init(db)

	// 连接池配置，防止 invalid connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// ⚠️ 关键：这里要改成 model.LostItem
	err = db.AutoMigrate(&model.Post{}, &model.User{}, &model.Session{}, &model.Message{},&model.Student{})
	if err != nil {
		panic("❌ 自动建表失败")
	}
	fmt.Println("✅ 数据库连接成功，表结构已同步！")
}

func main() {
	if len(config.JWTSecret) < 32 {
		panic("❌ 必须设置环境变量 JWT_SECRET（长度 ≥ 32 字节）")
	}
	// 4. 启动时先连数据库
	initDB()

	//__创建一个 Gin 引擎__，名字叫 `r`。它就是你的"服务器本体"，负责接收和处理所有 HTTP 请求。
	r := gin.New()
	//Recovery()防止程序崩溃
	r.Use(middleware.Recovery())
	//Logger()打日志
	r.Use(middleware.Logger())
	//CORS()	跨域 —— 让前端（不同端口）能正常访问你的后端
	r.Use(middleware.CORS())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	router.RegisterRoutes(r)
	r.Static("/uploads", config.UploadDir)

	addr := config.HTTPPort
	if addr != "" && addr[0] != ':' {
		addr = ":" + addr
	}
	r.Run(addr)
}
