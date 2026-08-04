package config

import "fmt"

// LoadYAML 按环境读 configs/config.{env}.yaml（阶段 H · 3.1）。
// 约定：文件提供默认值，同名环境变量覆盖文件。
//
// 依赖（实现时执行）：
//
//	go get github.com/spf13/viper
//
// 提示：
//   - SetConfigFile / AddConfigPath("configs")
//   - AutomaticEnv + SetEnvKeyReplacer
//   - Unmarshal 到 Config
func LoadYAML(env string) (Config, error) {
	// TODO(H1): 用 viper 加载 configs/config.<env>.yaml，再用 env 覆盖
	_ = env
	return Config{}, fmt.Errorf("TODO(H1): LoadYAML with viper not implemented")
}
