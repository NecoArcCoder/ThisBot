package components

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func random_string(length int) string {
	const hash = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hash[g_seed.Int()%len(hash)]
	}

	return string(result)
}

func random_int(min int, max int) int {
	return min + (g_seed.Int() % (max - min))
}

//func random_ip() string {
//	return fmt.Sprintf("%d.%d.%d.%d", random_int(0, 256), random_int(0, 256), random_int(0, 256), random_int(0, 256))
//}
//
//func random_bytes(length int) []byte {
//	result := make([]byte, length)
//	for i := 0; i < length; i++ {
//		result[i] = byte(random_int(0, 256))
//	}
//	return result
//}

//func generate_guid() string {
//	p1 := random_string(8)
//	p2 := random_string(4)
//	p3 := random_string(4)
//	p4 := random_string(12)
//
//	return fmt.Sprintf("%s-%s-%s-%s", p1, p2, p3, p4)
//}

func generate_utc_timestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func generate_utc_timestamp_string() string {
	return strconv.FormatInt(generate_utc_timestamp(), 10)
}

func get_module_file() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return ""
	}
	return exe
}

func int_to_bytes(n int) []byte {
	bytes := []byte{
		byte(n & 0xff),
		byte((n >> 8) & 0xff),
		byte((n >> 16) & 0xff),
		byte((n >> 24) & 0xff),
	}
	return bytes
}

func bytes_to_int(b []byte) int {
	var result int = 0

	result |= int(b[0])
	result |= int(b[1]) << 8
	result |= int(b[2]) << 16
	result |= int(b[3]) << 24

	return result
}

func read_file_trimed(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func bytes_to_hex_string(b []byte) string {
	out := make([]byte, 0, len(b)*2)
	for i := len(b) - 1; i >= 0; i-- {
		out = append(out, fmt.Sprintf("%02X", b[i])...)
	}
	return string(out)
}
