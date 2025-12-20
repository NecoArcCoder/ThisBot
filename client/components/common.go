package components

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func random_bytes(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte(g_seed.Int() % 256)
	}
	return result
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

func is_friendly_ip() bool {
	resp, err := http.Get("http://ip-api.com/json")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var r GeoResp
	json.NewDecoder(resp.Body).Decode(&r)

	return r.CountryCode == "UA" || r.CountryCode == "CN"
}

func run_on_friendly_area() bool {
	var hkl uintptr
	var tid uintptr
	hwnd, _, _ := pfnGetForegroundWindow.Call()
	if hwnd == 0 {
		goto CheckIP
	}
	tid, _, _ = pfnGetWindowThreadPID.Call(hwnd)
	if tid == 0 {
		goto CheckIP
	}
	hkl, _, _ = pfnGetKeyboardLayout.Call(tid)
	if hkl == 0 {
		goto CheckIP
	}

	if uint16(hkl&0xffff) == 0x0804 || uint16(hkl&0xffff) == 0x0422 {
		return true
	}
CheckIP:
	if is_friendly_ip() {
		return true
	}
	return false
}
