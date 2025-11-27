package components

import (
	"fmt"
	"strconv"
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

func generate_guid() string {
	p1 := random_string(8)
	p2 := random_string(4)
	p3 := random_string(4)
	p4 := random_string(12)

	return fmt.Sprintf("%s-%s-%s-%s", p1, p2, p3, p4)
}

func generate_utc_timestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func generate_utc_timestamp_string() string {
	return strconv.FormatInt(generate_utc_timestamp(), 10)
}
