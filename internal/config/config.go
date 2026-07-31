package config

import (
	"os"
)

// Config 应用配置（3.1：env 区分 local/staging/prod）。
type Config struct {
	AppEnv    string
	HTTPAddr  string
	MySQLDSN  string
	RedisAddr string
	JWTSecret string
}

// Load 从环境变量读取；缺省给本地练习默认值。
// TODO(3.1): 可改为 viper 读 yaml，env 覆盖文件。
func Load() Config {
	return Config{
		AppEnv:    getenv("APP_ENV", "local"),
		HTTPAddr:  getenv("HTTP_ADDR", ":8080"),
		MySQLDSN:  getenv("MYSQL_DSN", ""),
		RedisAddr: getenv("REDIS_ADDR", "127.0.0.1:6379"),
		JWTSecret: getenv("JWT_SECRET", "dev-secret-change-me"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
