package config

import (
	"fmt"
	"os"
)

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
		JWTSecret:    getEnv("JWT_SECRET", defaultJWTSecret),
		WechatAppID:  getEnv("WECHAT_APP_ID", ""),
		WechatSecret: getEnv("WECHAT_SECRET", ""),
	}
}

// Validate checks production safety. Fatals if critical secrets are unset or still at defaults.
func (c *Config) Validate() {
	env := os.Getenv("APP_ENV")
	if env != "production" {
		return
	}

	issues := []string{}
	if c.JWTSecret == defaultJWTSecret {
		issues = append(issues, "JWT_SECRET is still the default value")
	}
	if c.DBPassword == "spark123" {
		issues = append(issues, "DB_PASSWORD is still the default value")
	}
	if len(issues) > 0 {
		for _, s := range issues {
			fmt.Fprintf(os.Stderr, "[CONFIG] FATAL: %s\n", s)
		}
		fmt.Fprintf(os.Stderr, "[CONFIG] Set APP_ENV=development to skip production checks.\n")
		os.Exit(1)
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

const defaultJWTSecret = "spark-dev-secret-change-in-prod"
