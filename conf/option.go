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

package conf

import (
	"reflect"

	"github.com/antigloss/go/conf/store"
)

// WithStores 设置 ConfigParser 的配置数据来源
func WithStores(stores ...store.Store) option {
	return func(o *options) {
		o.stores = stores
	}
}

// WithTagName 设置把配置数据反序列化到结构体时，使用的 TagName ，默认是 mapstructure
func WithTagName(tag string) option {
	return func(o *options) {
		o.tagName = tag
	}
}

// DecodeHook 配置数据解码器
type DecodeHook struct {
	// 自定义数据类型
	Type reflect.Type
	// 自定义数据类型解码器。入参为配置原始字符串，返回为解码后的结果
	Decode func(data string) (interface{}, error)
}

// WithDecodeHooks 设置自定义类型数据解码器
func WithDecodeHooks(hooks ...DecodeHook) option {
	return func(o *options) {
		o.hooks = hooks
	}
}

type option func(opts *options)

type options struct {
	stores  []store.Store
	tagName string
	hooks   []DecodeHook
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
