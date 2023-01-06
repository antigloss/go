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

// ConfigPath holds path to the local configuration file/directory
type ConfigPath struct {
	Path      string // path to the local configuration file/directory. If it's a directory, then all files under the directory will also be read, excluding those start with a dot (.)
	Recursive bool   // true for reading files under `Path` recursively, false for reading only the files right under directory `Path`
}

// WithConfigPaths sets paths to local configuration files
func WithConfigPaths(paths ...ConfigPath) option {
	return func(o *options) {
		o.paths = paths
	}
}

// WithTemplateData sets template data source.
// Will use configurations from `tData` to replace templates in the configurations from local files
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
