package config

import "os"

func getEnv(key string, v string) string {
	value := os.Getenv(key)

	if value == "" {
		return v
	}
	return value

}

var UploadDir = getEnv("UPLOAD_DIR", "./uploads")
var BaseURL = getEnv("BASE_URL", "http://localhost:8000")
var JWTSecret = getEnv("JWT_SECRET", "")
var DatabaseDSN = getEnv("DATABASE_DSN", "root:123456@tcp(127.0.0.1:3306)/lost_found_db?charset=utf8mb4&parseTime=True&loc=Local")
var HTTPPort = getEnv("PORT", "8000")

// SMTP 邮件服务。默认全空 = "没配"，此时 main.go 会继续用打印版的假发送器，
// 这样本地开发不配邮箱也能跑通整条验证码链路。
// 注意：SMTP_PASS 是邮箱的【授权码】，不是登录密码。
var SMTPHost = getEnv("SMTP_HOST", "")
var SMTPPort = getEnv("SMTP_PORT", "587")
var SMTPUser = getEnv("SMTP_USER", "")
var SMTPPass = getEnv("SMTP_PASS", "")
var SMPT_Host = getEnv("","")
var SMPT_Port = getEnv("","587")
var SMPT_User = getEnv("","")
var SMPT_Pass = getEnv("","")