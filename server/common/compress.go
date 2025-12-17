package common

import (
	"bytes"
	"compress/zlib"
	"io"
)

// After the test, only the data size is more than 54 bytes will work
func ZlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	w, err := zlib.NewWriterLevel(&buf, zlib.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ZlibDecompress(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
