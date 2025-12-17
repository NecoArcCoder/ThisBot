package utils

import (
	"ThisBot/common"
	"bufio"
	"crypto/rand"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func RandomString(length int) string {
	const hash = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hash[common.Seed.Int()%len(hash)]
	}

	return string(result)
}

func RandomInt(min int, max int) int {
	return min + (common.Seed.Int() % (max - min))
}

func GenerateUtcTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func GenerateUtcTimestampString() string {
	return strconv.FormatInt(GenerateUtcTimestamp(), 10)
}

func GenerateRandomBytes(length int) []byte {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil
	}
	return b
}

func ReadJson(r *http.Request, out any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()

	return json.Unmarshal(body, out)
}

func ReadFromIO() string {
	command, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	command = strings.ToLower(strings.TrimSpace(command))
	return command
}

func IsValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

func IsValidDomain(s string) bool {
	var domainRegex = regexp.MustCompile(`^(?i)[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$`)
	return domainRegex.MatchString(s)
}

func IsLegalURLOrIP(url0 string) bool {
	if IsValidDomain(url0) || url0 == "localhost" {
		return true
	}
	//ret, err := url.Parse(url0)
	//if err == nil && (ret.Host != "" && ret.Scheme != "") {
	//	return true
	//}
	return IsValidIP(url0)
}

func IntToBytes(n int) []byte {
	bytes := []byte{
		byte(n & 0xff),
		byte((n >> 8) & 0xff),
		byte((n >> 16) & 0xff),
		byte((n >> 24) & 0xff),
	}
	return bytes
}

func BytesToInt(b []byte) int {
	var result int = 0

	result |= int(b[0])
	result |= int(b[1]) << 8
	result |= int(b[2]) << 16
	result |= int(b[3]) << 24

	return result
}

func TimestampStringToMySqlDateTime(timestamp string) string {
	unixTimestamp, _ := strconv.ParseInt(timestamp, 10, 64)
	utcTime := time.UnixMilli(unixTimestamp).UTC()
	return utcTime.Format("2006-01-02 15:04:05")
}
