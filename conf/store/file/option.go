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

package file

import "github.com/antigloss/go/conf/tdata"

// ConfigPath 配置文件路径
type ConfigPath struct {
	Path      string // 配置文件路径，可以是具体的文件，也可以是目录。如果路径是目录，则会读取目录下的所有文件（不包括 . 开头的文件）
	Recursive bool   // true 表示递归读取所有子目录的文件，false 则只读取 Path 指定目录的文件
}

// WithConfigPaths 设置配置文件路径。如果路径是目录，则会读取目录下的所有文件（不包括 . 开头的文件）
func WithConfigPaths(paths ...ConfigPath) option {
	return func(o *options) {
		o.paths = paths
	}
}

// WithTemplateData 开启模板替换功能，使用 tData 替换配置中的模板参数
func WithTemplateData(tData tdata.TemplateData) option {
	return func(o *options) {
		o.tData = tData
	}
}

type option func(options *options)

type options struct {
	paths []ConfigPath
	tData tdata.TemplateData
}

func (o *options) apply(opts ...option) {
	for _, opt := range opts {
		opt(o)
	}
}
