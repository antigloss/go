package store

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigTypeDefault = "properties" // 默认配置格式
	ConfigTypeJSON    = "json"       // JSON
	ConfigTypeYAML    = "yaml"       // YAML
	ConfigTypeYML     = "yml"        // YAML
	ConfigTypeEnv     = "env"        // 环境变量
)

// ConfigType 使用文件名后缀作为配置格式，如：properties、xml、yml、yaml、json 等。
//   - 如果没有后缀名，则默认为 properties
//   - 如果不支持后缀名指定的配置格式，则返回 error
func ConfigType(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if len(ext) > 1 {
		ext = ext[1:]
		for _, e := range viper.SupportedExts {
			if e == ext {
				return ext, nil
			}
		}
		return "", fmt.Errorf("unsupported configuration format: %s", ext)
	}
	return ConfigTypeDefault, nil
}
