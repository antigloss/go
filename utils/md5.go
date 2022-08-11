/*
 *
 * sync - Synchronization facilities.
 * Copyright (C) 2022 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// MD5File returns the MD5 checksum of the file contents.
//
//	`filepath` - Path to the file
func MD5File(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	md := md5.New()
	_, err = io.Copy(md, file)
	if err == nil {
		return md.Sum(nil), nil
	}

	return nil, err
}

// MD5FileString returns the MD5 checksum of the file contents, in lowercase hex string.
//
//	`filepath` - Path to the file
func MD5FileString(filepath string) (string, error) {
	md, err := MD5File(filepath)
	if err == nil {
		return hex.EncodeToString(md), nil
	}
	return "", err
}
