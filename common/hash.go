package common

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func GetFileHash(path string) (string, error) {
	if file, err := os.Open(path); err == nil {
		defer file.Close()
		hash := md5.New()
		if _, err := io.Copy(hash, file); err == nil {
			hashBytes := hash.Sum(nil)
			return hex.EncodeToString(hashBytes), nil
		} else {
			return "", err
		}
	} else {
		return "", err
	}
}
