/*
 *
 * Copyright (C) 2023 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package conf is a framework for reading configurations from variety of configuration Stores, such as ENV, files, and Apollo.
package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"

	"github.com/antigloss/go/conf/store"
)

// New creates a ConfigParser object
//   - T is the struct for unmarshalling configuration data
func New[T any](opts ...option) *ConfigParser[T] {
	isSlice := false
	var t T
	ty := reflect.TypeOf(t)
	if ty.Kind() == reflect.Slice {
		isSlice = true
	}

	c := &ConfigParser[T]{
		isSlice:   isSlice,
		viper:     viper.New(),
		changesCh: make(chan *store.ConfigChanges, 20),
		unwatchCh: make(chan int),
	}
	c.opts.apply(opts...)
	return c
}

// ConfigParser is a configuration data parser. It supports variety of configuration Stores, mainstream configuration formats, watching, and templates
//   - `T` is the struct for unmarshalling configuration data
type ConfigParser[T any] struct {
	opts      options
	isSlice   bool
	sliceLen  int
	viper     *viper.Viper
	changesCh chan *store.ConfigChanges
	unwatchCh chan int
	watchOnce sync.Once
}

// Parse reads configuration data from all Stores, then unmarshal it to `T`.
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
			err = c.transformArray(&cont)
			if err != nil {
				return nil, err
			}

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

// Watch watches configuration changes from all Stores, unmarshal the latest configuration data into `T`, then notify the caller via `cb`
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
					e := c.transformArray(&changes.Config)
					if e != nil {
						continue
					}

					c.viper.SetConfigType(changes.Config.Type)
					e = c.viper.MergeConfig(bytes.NewReader(changes.Config.Content))
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

// Unwatch stops watching
func (c *ConfigParser[T]) Unwatch() {
	for _, store := range c.opts.stores {
		store.Unwatch()
	}
	close(c.unwatchCh)
}

func (c *ConfigParser[T]) initDefaultValues(v reflect.Value) error {
	if v.Kind() == reflect.Struct {
		m := map[string]interface{}{}
		c.getDefaultValues(v.Type(), m)
		c.viper.SetConfigType(store.ConfigTypeYAML)
		return c.viper.MergeConfigMap(m)
	}
	return nil
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

// transformArray 把数组格式的配置，转换成对象格式
func (c *ConfigParser[T]) transformArray(cont *store.ConfigContent) error {
	if !c.isSlice {
		return nil
	}

	var err error
	var s []interface{}

	switch cont.Type {
	case store.ConfigTypeJSON:
		err = json.Unmarshal(cont.Content, &s)
	case store.ConfigTypeYAML, store.ConfigTypeYML:
		err = yaml.Unmarshal(cont.Content, &s)
	default:
		err = fmt.Errorf("unsupported config format: %s", cont.Type)
	}

	if err != nil {
		return err
	}

	m := map[int]interface{}{}
	for k, v := range s {
		m[k] = v
	}

	bs, err := json.Marshal(m)
	if err != nil {
		return err
	}

	c.sliceLen = len(s)
	cont.Type = store.ConfigTypeJSON
	cont.Content = bs
	return nil
}

func (c *ConfigParser[T]) unmarshal(t *T) error {
	if !c.isSlice {
		return c.viper.Unmarshal(t, func(config *mapstructure.DecoderConfig) {
			if c.opts.tagName != "" {
				config.TagName = c.opts.tagName
			}
		}, viper.DecodeHook(decodeHook(c.opts.hook)))
	}

	ty := reflect.TypeOf(*t)
	v := reflect.ValueOf(*t)
	for i := 0; i < c.sliceLen; i++ {
		elem := reflect.New(ty.Elem())
		err := c.viper.UnmarshalKey(strconv.Itoa(i), elem.Interface(), func(config *mapstructure.DecoderConfig) {
			if c.opts.tagName != "" {
				config.TagName = c.opts.tagName
			}
		}, viper.DecodeHook(decodeHook(c.opts.hook)))
		if err != nil {
			return err
		}

		v = reflect.Append(v, elem.Elem())
	}

	*t = v.Interface().(T)
	return nil
}
