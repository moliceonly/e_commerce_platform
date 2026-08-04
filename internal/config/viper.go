package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func LoadYAML(env string) (Config, error) {
	// TODO(H1): 用 viper 加载 configs/config.<env>.yaml，再用 env 覆盖
	if env == "" {
		env = "local"
	}

	v := viper.New()
	v.SetConfigFile(fmt.Sprintf("configs/config.%s.yaml", env))
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("app_env", "APP_ENV")
	_ = v.BindEnv("http_addr", "HTTP_ADDR")
	_ = v.BindEnv("mysql_dsn", "MYSQL_DSN")
	_ = v.BindEnv("redis_addr", "REDIS_ADDR")
	_ = v.BindEnv("jwt_secret", "JWT_SECRET")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.AppEnv == "" {
		cfg.AppEnv = env
	}
	return cfg, nil
}
