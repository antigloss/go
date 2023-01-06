package env

import (
	"bytes"
	"fmt"
	"os"

	"github.com/antigloss/go/conf/store"
)

// New 创建从 os.Environ() 读取配置的 Store 对象。
func New(opts ...option) store.Store {
	a := &envStore{}
	a.opts.apply(opts...)
	return a
}

type envStore struct {
	opts options
}

// Load 加载配置
func (a *envStore) Load() ([]store.ConfigContent, error) {
	buf := bytes.NewBuffer(nil)
	for _, env := range os.Environ() {
		fmt.Fprintln(buf, env)
	}

	contents := make([]store.ConfigContent, 1)
	contents[0].Type = store.ConfigTypeEnv
	contents[0].Content = buf.Bytes()

	if a.opts.tData != nil {
		var err error
		contents[0].Content, err = a.opts.tData.Replace(contents[0].Content)
		if err != nil {
			return nil, fmt.Errorf("%s: ENV", err.Error())
		}
	}

	return contents, nil
}

// Watch 监听配置变化。暂时不支持该操作，直接返回 nil
func (a *envStore) Watch(ch chan<- *store.ConfigChanges) error {
	return nil
}

// Unwatch 取消监听
func (a *envStore) Unwatch() {
}
