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
