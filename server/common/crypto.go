package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// Base64 encryption
func Base64Enc(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64 decryption
func Base64Dec(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

// Sha256
func Sha256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Hmac based on sha256
func HmacSha256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
