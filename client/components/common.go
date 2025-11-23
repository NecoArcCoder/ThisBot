package components

func random_string(length int) string {
	const hash = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hash[seed.Int()%len(hash)]
	}

	return string(result)
}

func random_int(min int, max int) int {
	return min + (seed.Int() % (max - min))
}
