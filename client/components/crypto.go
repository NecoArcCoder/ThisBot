package components

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/chacha20"
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

func enc_chacha20(key, plain []byte) []byte {
	nonce := make([]byte, chacha20.NonceSize)
	_, err := rand.Read(nonce)
	if err != nil {
		return nil
	}
	cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return nil
	}

	cipher_text := make([]byte, len(plain))
	cipher.XORKeyStream(cipher_text, plain)

	return append(nonce, cipher_text...)
}

func dec_chacha20(key, nonceCipher []byte) []byte {
	nonce := nonceCipher[:chacha20.NonceSize]
	text := nonceCipher[chacha20.NonceSize:]

	cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		return nil
	}
	plain := make([]byte, len(text))

	cipher.XORKeyStream(plain, text)
	return plain
}
