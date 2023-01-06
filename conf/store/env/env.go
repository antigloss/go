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

package env

import (
	"bytes"
	"fmt"
	"os"

	"github.com/antigloss/go/conf/store"
)

// New creates a Store object to read configurations from os.Environ()
func New(opts ...option) store.Store {
	a := &envStore{}
	a.opts.apply(opts...)
	return a
}

type envStore struct {
	opts options
}

// Load reads configurations
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

// Watch watches configuration changes. Not yet supported
func (a *envStore) Watch(ch chan<- *store.ConfigChanges) error {
	return nil
}

// Unwatch stops watching
func (a *envStore) Unwatch() {
}
