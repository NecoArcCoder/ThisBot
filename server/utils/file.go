package utils

import "os"

// Check if a file is exist or not
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
