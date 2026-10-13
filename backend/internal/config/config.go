package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config คือค่าที่ใช้รันแอป อ่านจาก environment / ไฟล์ .env
type Config struct {
	HTTPPort string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
}

func Load() Config {
	// รันจากโฟลเดอร์ backend → .env อยู่โฟลเดอร์บน
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	return Config{
		HTTPPort: env("HTTP_PORT", "8080"),
		DBHost:   env("MYSQL_HOST", "127.0.0.1"),
		DBPort:   env("MYSQL_PORT", "3306"),
		DBUser:   env("MYSQL_USER", "root"),
		DBPass:   env("MYSQL_PASSWORD", ""),
		DBName:   env("MYSQL_DATABASE", "trackclient"),
	}
}

// DSN คือที่อยู่ที่ driver MySQL ใช้ต่อ (Data Source Name)
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
