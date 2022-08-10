/*
 *
 * http_utils - Handy HTTP utilities.
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

package http_utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Get sends an http GET request and returns the response body as string
func Get(cli *http.Client, url string) (string, error) {
	rsp, err := cli.Get(url)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	cont, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(cont), nil
}

// GetBytes sends an http GET request and returns the response body as []byte
func GetBytes(cli *http.Client, url string) ([]byte, error) {
	rsp, err := cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	cont, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	return cont, nil
}

// Download downloads the file from `url` and saves it to `dstFilepath`
func Download(cli *http.Client, url, dstFilepath string) error {
	rsp, err := cli.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	tmpFile := dstFilepath + "-_v.~v~tmp^_^"
	file, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporal file")
	}
	defer file.Close()

	_, err = io.Copy(file, rsp.Body)
	if err != nil {
		os.Remove(tmpFile)
		return err
	}

	err = os.Rename(tmpFile, dstFilepath)
	if err == nil {
		return nil
	}

	os.Remove(tmpFile)
	return err
}
