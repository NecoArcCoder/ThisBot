package components

import (
	"crypto/hmac"
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
func sha256_hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Hmac based on sha256
func hmac_sha256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// Create sha256-based HMAC
func create_sign(token string, guid string, timestamp string) []byte {
	bytToken, _ := base64_dec(token)
	data := []byte(guid + timestamp)
	return hmac_sha256(bytToken, data)
}
