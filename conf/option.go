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

// DecodeHook decoder for a specified data type
type DecodeHook struct {
	// Data type
	Type reflect.Type
	// Decoder for decoding raw configuration data into `Type`.
	// It returns the decoded value as interface{} on success, otherwise, an error is returned.
	Decode func(data string) (interface{}, error)
}

// WithDecodeHooks sets user-defined decoders
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
