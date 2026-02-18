package config

import (
	"os"
)

// Config struct สำหรับเก็บการตั้งค่าทั้งหมด
type Config struct {
	AppAddr   string
	AppEnv    string
	JWTSecret string
	DBDSN     string
}

// Load ฟังก์ชันนี้ใช้โหลดค่าจาก .env และคืนค่า Config
func Load() Config {
	return Config{
		AppAddr:   getEnv("APP_ADDR", ":8080"),
		AppEnv:    getEnv("APP_ENV", "dev"),
		JWTSecret: mustGetEnv("JWT_SECRET"),
		DBDSN:     mustGetEnv("DB_DSN"),
	}
}

// getEnv ใช้ดึงค่าจาก ENV ถ้าไม่เจอจะคืนค่าที่กำหนด
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// mustGetEnv ใช้ดึงค่า ENV ที่จำเป็น ถ้าไม่เจอจะ panic
func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing required environment variable: " + key)
	}
	return value
}
