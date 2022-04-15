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
