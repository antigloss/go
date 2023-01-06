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

package apollo

import (
	"io/fs"
	"os"

	yaml "gopkg.in/yaml.v3"
)

// NewLocalConfig 读取本地配置，从中获取 Apollo 秘钥，并使用本地配置覆盖 Apollo 上的配置
func NewLocalConfig(o ...localOption) (*localConfig, error) {
	var opts localOptions
	opts.apply(o...)

	c := &localConfig{conf: map[string]map[string]string{}}

	cont, err := os.ReadFile(opts.cfgPath)
	if err == nil {
		if err = yaml.Unmarshal(cont, &c.conf); err != nil {
			return nil, err
		}
	} else {
		if _, ok := err.(*fs.PathError); !ok {
			return nil, err
		}
	}

	return c, nil
}

// WithLocalConfigPath 设置本地配置覆盖文件的路径
func WithLocalConfigPath(path string) localOption {
	return func(opts *localOptions) {
		opts.cfgPath = path
	}
}

type localConfig struct {
	conf map[string]map[string]string
}

type localOption func(opts *localOptions)

type localOptions struct {
	cfgPath string
}

func (c *localOptions) apply(opts ...localOption) {
	for _, opt := range opts {
		opt(c)
	}

	if c.cfgPath == "" {
		c.cfgPath = "configs.yaml"
	}
}
