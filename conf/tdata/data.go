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

// Package tdata implements TemplateData from which data can be used to replace templates from other Stores.
package tdata

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

// New creates a TemplateData object which supports the following user-defined functions:
//
//   - env KEY      replace `env KEY` with the value of `KEY` read from ENV
//   - hostname     replace `hostname` with the value of os.Hostname()
//   - value KEY    replace `value KEY` with the value of `KEY` read from Stores assigned to the TemplateData object
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

// TemplateData provides data for replacing templates
type TemplateData interface {
	Replace(tpl []byte) ([]byte, error) // use data from TemplateData to replace templates in `tpl`
}

type templateData struct {
	opts  options
	viper *viper.Viper
}

// Replace uses data from TemplateData to replace templates in `tpl`
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
