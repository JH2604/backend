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
