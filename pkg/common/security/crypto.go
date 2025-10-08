package security

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
)

var (
	encryptionKey []byte
	encryptionIV  []byte
)

// Initialize loads Unity CBC encryption keys from environment variables
func Initialize() error {
	// Load Unity CBC encryption key from environment
	keyStr := os.Getenv("UNITY_CBC_ENCRYPTION_KEY")
	if keyStr == "" {
		return errors.New("UNITY_CBC_ENCRYPTION_KEY environment variable not set")
	}
	encryptionKey = []byte(keyStr)

	// Load Unity CBC encryption IV from environment
	ivStr := os.Getenv("UNITY_CBC_ENCRYPTION_IV")
	if ivStr == "" {
		return errors.New("UNITY_CBC_ENCRYPTION_IV environment variable not set")
	}
	encryptionIV = []byte(ivStr)

	// Validate key and IV lengths
	if len(encryptionKey) != 16 {
		return errors.New("Unity CBC encryption key must be 16 bytes")
	}
	if len(encryptionIV) != 16 {
		return errors.New("Unity CBC encryption IV must be 16 bytes")
	}

	log.Println("Unity CBC encryption system initialized with environment variables")
	return nil
}

// EncryptToken encrypts a token using AES-CBC with fixed IV
func EncryptToken(token string) (string, error) {
	// Create AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Pad the token to a multiple of the block size
	padding := aes.BlockSize - (len(token) % aes.BlockSize)
	padtext := make([]byte, len(token)+padding)
	copy(padtext, token)
	for i := len(token); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}

	// Encrypt the padded token
	ciphertext := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, encryptionIV)
	mode.CryptBlocks(ciphertext, padtext)

	// Encode the ciphertext in Base64
	encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

	return encodedCiphertext, nil
}

// DecryptToken decrypts a token that was encrypted with EncryptToken
func DecryptToken(encryptedToken string) (string, error) {
	// Decode Base64 ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", errors.New("failed to decode ciphertext")
	}

	// Create AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	// Decrypt the ciphertext
	mode := cipher.NewCBCDecrypter(block, encryptionIV)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize {
		return "", errors.New("invalid padding")
	}

	return string(plaintext[:len(plaintext)-padding]), nil
}

// DecryptRequestCBC decrypts a request using Unity's CBC approach
func DecryptRequestCBC(body []byte) ([]byte, error) {
	decryptedData, err := DecryptToken(string(body))
	if err != nil {
		return nil, err
	}

	return []byte(decryptedData), nil
}

// EncryptResponseCBC encrypts a response using Unity's CBC approach
func EncryptResponseCBC(data interface{}) ([]byte, error) {
	// Convert response to JSON
	responseJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Encrypt using Unity's CBC approach
	encryptedData, err := EncryptToken(string(responseJSON))
	if err != nil {
		return nil, err
	}

	// Return as plain string
	return []byte(encryptedData), nil
}
