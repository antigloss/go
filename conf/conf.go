package conf

import (
	"bytes"
	"reflect"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"github.com/antigloss/go/conf/store"
)

// New 创建 ConfigParser 对象
func New[T any](opts ...option) *ConfigParser[T] {
	c := &ConfigParser[T]{
		viper:     viper.New(),
		changesCh: make(chan *store.ConfigChanges, 20),
		unwatchCh: make(chan int),
	}
	c.opts.apply(opts...)
	return c
}

// ConfigParser 配置解析器。支持多种配置存储方式、多种配置格式、动态更新、模板替换
//
//	T - 承载配置解析结果的结构体
type ConfigParser[T any] struct {
	opts      options
	viper     *viper.Viper
	changesCh chan *store.ConfigChanges
	unwatchCh chan int
	watchOnce sync.Once
}

// Parse 从各个 Store 中读取配置，然后反序列化到 T 里面
func (c *ConfigParser[T]) Parse() (*T, error) {
	var t T

	err := c.initDefaultValues(reflect.ValueOf(t))
	if err != nil {
		return nil, err
	}

	for _, store := range c.opts.stores {
		contents, err := store.Load()
		if err != nil {
			return nil, err
		}

		for _, cont := range contents {
			c.viper.SetConfigType(cont.Type)
			err = c.viper.MergeConfig(bytes.NewReader(cont.Content))
			if err != nil {
				return nil, err
			}
		}
	}

	err = c.unmarshal(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// Watch 监听各个 Store 的配置变更，把最新的配置反序列化到 T 里面，然后通过回调函数告知
func (c *ConfigParser[T]) Watch(cb func(cfg *T, changes []store.ConfigChange)) error {
	var err error

	c.watchOnce.Do(func() {
		for _, store := range c.opts.stores {
			if err = store.Watch(c.changesCh); err != nil {
				return
			}
		}

		go func() {
			for {
				select {
				case changes := <-c.changesCh:
					c.viper.SetConfigType(changes.Config.Type)
					e := c.viper.MergeConfig(bytes.NewReader(changes.Config.Content))
					if e != nil {
						continue
					}

					var t T
					e = c.unmarshal(&t)
					if e != nil {
						continue
					}

					cb(&t, changes.Changes)
				case <-c.unwatchCh:
					return
				}
			}
		}()
	})

	return err
}

// Unwatch 取消监听
func (c *ConfigParser[T]) Unwatch() {
	for _, store := range c.opts.stores {
		store.Unwatch()
	}
	close(c.unwatchCh)
}

func (c *ConfigParser[T]) initDefaultValues(v reflect.Value) error {
	m := map[string]interface{}{}
	c.getDefaultValues(v.Type(), m)
	c.viper.SetConfigType(store.ConfigTypeYAML)
	return c.viper.MergeConfigMap(m)
}

func (c *ConfigParser[T]) getDefaultValues(t reflect.Type, m map[string]interface{}) {
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tagName := ft.Tag.Get(c.opts.tagName)
		if tagName == "" {
			tagName = strings.ToLower(ft.Name)
		}

		fv := t.Field(i).Type
		if fv.Kind() == reflect.Pointer {
			fv = fv.Elem()
		}
		if fv.Kind() != reflect.Struct {
			defVal := ft.Tag.Get("default")
			if defVal != "" {
				m[tagName] = defVal
			}
			continue
		}

		mm := map[string]interface{}{}
		c.getDefaultValues(fv, mm)
		if len(mm) > 0 {
			m[tagName] = mm
		}
	}
}

func (c *ConfigParser[T]) unmarshal(t *T) error {
	return c.viper.Unmarshal(t, func(config *mapstructure.DecoderConfig) {
		if c.opts.tagName != "" {
			config.TagName = c.opts.tagName
		}
	}, viper.DecodeHook(decodeHook(c.opts.hooks)))
}
