package utils

import (
	"io"
	"os"
)

// Check if a file is exist or not
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func WriteBinary(path string, payload []byte) bool {
	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(payload)
	if err != nil {
		return false
	}
	return true
}

func ReadBinary(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, 0, 1024*1024)
	tmp := make([]byte, 4096)

	for {
		n, err := f.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return buf, nil
}
