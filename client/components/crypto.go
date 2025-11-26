package components

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// Base64 encryption
func base64_enc(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64 decryption
func base64_dec(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// Sha256
func Sha256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
