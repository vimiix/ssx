package util

import (
	"errors"
	"io/fs"
	"os"
)

// FileExists check given filename if exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
		panic(err)
	}
	return true
}
