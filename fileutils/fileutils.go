/*
 *
 * fileutils - Handy file utilities.
 * Copyright (C) 2018 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

// Package fileutils provides some handy file utilities.
package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDirectory copies a directory from src to dst recursively.
func CopyDirectory(src, dst string) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcFileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", src)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcFileInfo.Mode())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if !entry.IsDir() {
			srcFile, err := os.Open(srcPath)
			if err != nil {
				return err
			}

			dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Type())
			if err != nil {
				return err
			}

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}
		} else {
			err = CopyDirectory(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ClearDirectory removes all files and directories under `dir` recursively.
func ClearDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err = os.RemoveAll(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
