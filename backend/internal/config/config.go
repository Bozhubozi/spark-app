package config

import "os"

type Config struct {
	ServerPort   string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	RedisAddr    string
	RedisPass    string
	JWTSecret    string
	WechatAppID  string
	WechatSecret string
}

func Load() *Config {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "spark"),
		DBPassword:   getEnv("DB_PASSWORD", "spark123"),
		DBName:       getEnv("DB_NAME", "spark"),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:    getEnv("REDIS_PASS", ""),
		JWTSecret:    getEnv("JWT_SECRET", "spark-dev-secret-change-in-prod"),
		WechatAppID:  getEnv("WECHAT_APP_ID", ""),
		WechatSecret: getEnv("WECHAT_SECRET", ""),
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost + " port=" + c.DBPort +
		" user=" + c.DBUser + " password=" + c.DBPassword +
		" dbname=" + c.DBName + " sslmode=disable TimeZone=Asia/Shanghai"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
