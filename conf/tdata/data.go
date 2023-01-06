package tdata

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

// New 创建 TemplateData 对象。支持了 3 个自定义函数：
//
//   - env KEY      替换为环境变量 KEY 的值
//   - hostname     替换为 hostname
//   - value KEY    替换为 Store 中和 KEY 同名的值
func New(opts ...option) (TemplateData, error) {
	t := &templateData{viper: viper.New()}
	t.opts.apply(opts...)

	for _, store := range t.opts.stores {
		contents, err := store.Load()
		if err != nil {
			return nil, err
		}

		for _, cont := range contents {
			t.viper.SetConfigType(cont.Type)
			err = t.viper.MergeConfig(bytes.NewReader(cont.Content))
			if err != nil {
				return nil, err
			}
		}
	}

	return t, nil
}

// TemplateData 为配置中的模板参数（text/template）提供替换数据
type TemplateData interface {
	Replace(tpl []byte) ([]byte, error) // 用 TemplateData 中的数据，替换 tpl 中的模板参数
}

type templateData struct {
	opts  options
	viper *viper.Viper
}

// Replace 用 templateData 中的数据，替换 tpl 中的模板参数
func (t *templateData) Replace(tpl []byte) ([]byte, error) {
	tp := template.New("")
	tp.Funcs(map[string]any{
		"env":      os.Getenv,
		"hostname": hostname,
		"value":    t.value,
	})

	tp, err := tp.Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	result := bytes.NewBuffer(nil)
	err = tp.Execute(result, nil)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (t *templateData) value(key string) string {
	if v := t.viper.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	return ""
}

func hostname() string {
	name, _ := os.Hostname()
	return name
}
