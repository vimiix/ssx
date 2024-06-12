package file

import (
	"io"
	"os"
)

// CopyFile copies the contents of src to dst
func CopyFile(src, dst string, perm os.FileMode) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	tf, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer tf.Close()
	_, err = io.Copy(tf, sf)
	return err
}

// IsExist check given path if exists
func IsExist(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
