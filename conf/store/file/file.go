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

// Package file implements a Store client for reading and watching configurations from local files.
package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antigloss/go/conf/store"
)

// New creates a Store object to read configurations from local files
func New(opts ...option) store.Store {
	a := &fileStore{}
	a.opts.apply(opts...)
	return a
}

type fileStore struct {
	opts options
}

// Load reads configurations
func (a *fileStore) Load() ([]store.ConfigContent, error) {
	paths, err := a.calculateFilePaths()
	if err != nil {
		return nil, err
	}

	contents := make([]store.ConfigContent, len(paths))
	for i, p := range paths {
		contents[i].Type, err = store.ConfigType(p)
		if err != nil {
			return nil, err
		}

		contents[i].Content, err = os.ReadFile(p)
		if err != nil {
			return nil, err
		}

		if a.opts.tData != nil {
			contents[i].Content, err = a.opts.tData.Replace(contents[i].Content)
			if err != nil {
				return nil, fmt.Errorf("%s: %s", err.Error(), p)
			}
		}
	}
	return contents, nil
}

// Watch watches configuration changes. Not yet supported
func (a *fileStore) Watch(ch chan<- *store.ConfigChanges) error {
	return nil
}

// Unwatch stops watching
func (a *fileStore) Unwatch() {
}

func (a *fileStore) calculateFilePaths() ([]string, error) {
	var paths []string

	for _, p := range a.opts.paths {
		f, err := os.Stat(p.Path)
		if err != nil {
			return nil, err
		}

		if f.IsDir() {
			ps, e := readDir(p.Path, p.Recursive)
			if e != nil {
				return nil, err
			}
			paths = append(paths, ps...)
		} else {
			paths = append(paths, p.Path)
		}
	}

	return paths, nil
}

func readDir(dir string, recursive bool) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if !file.IsDir() {
			paths = append(paths, filepath.Join(dir, file.Name()))
			continue
		}

		if !recursive {
			continue
		}

		ps, e := readDir(filepath.Join(dir, file.Name()), true)
		if e != nil {
			return nil, e
		}
		paths = append(paths, ps...)
	}
	return paths, nil
}
