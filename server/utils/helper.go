package utils

import (
	"ThisBot/common"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
