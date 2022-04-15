package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// MD5File returns the MD5 checksum of the file contents.
//  `filepath` - Path to the file
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
//  `filepath` - Path to the file
func MD5FileString(filepath string) (string, error) {
	md, err := MD5File(filepath)
	if err == nil {
		return hex.EncodeToString(md), nil
	}
	return "", err
}
