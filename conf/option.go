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

// WithStores sets Stores to ConfigParser
func WithStores(stores ...store.Store) option {
	return func(o *options) {
		o.stores = stores
	}
}

// WithTagName sets tag name used when unmarshalling data to the configuration struct. Default is mapstructure
func WithTagName(tag string) option {
	return func(o *options) {
		o.tagName = tag
	}
}

// DecodeHook decodes `data` to an object of type `to`.
// It returns the decoded value as interface{} on success, otherwise, an error is returned.
type DecodeHook func(to reflect.Type, data string) (interface{}, error)

// WithDecodeHook sets a user-defined decoder
func WithDecodeHook(hook DecodeHook) option {
	return func(o *options) {
		o.hook = hook
	}
}

type option func(opts *options)

type options struct {
	stores  []store.Store
	tagName string
	hook    DecodeHook
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
