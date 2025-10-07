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

// Unity CBC encryption keys loaded from environment variables
// WARNING: Fixed IV is a security vulnerability - should be random for each encryption
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

// =============================================================================
// UNITY TEAM CBC IMPLEMENTATION (as requested)
// WARNING: This implementation has security vulnerabilities:
// 1. Fixed IV allows pattern analysis attacks
// 2. No integrity protection - data can be tampered with
// 3. No authentication - no way to verify data origin
// 4. No replay protection - old requests can be replayed
// 5. CBC mode is malleable - attackers can modify ciphertext
// =============================================================================

// EncryptToken encrypts a token using AES-CBC with fixed IV (Unity team's approach)
// WARNING: This is INSECURE due to fixed IV and lack of integrity protection
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
// WARNING: This is INSECURE due to fixed IV and lack of integrity protection
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

// =============================================================================
// SIMPLE CBC REQUEST/RESPONSE FUNCTIONS (Unity-compatible)
// These functions use the Unity team's CBC approach for full request/response
// =============================================================================

// DecryptRequestCBC decrypts a request using Unity's CBC approach
// WARNING: This bypasses all security features (HMAC, timestamp, nonce)
func DecryptRequestCBC(body []byte) ([]byte, error) {
	// For Unity's simple approach, we expect just the encrypted JSON directly
	// No SecurePackage wrapper, no HMAC, no timestamp validation

	decryptedData, err := DecryptToken(string(body))
	if err != nil {
		return nil, err
	}

	return []byte(decryptedData), nil
}

// EncryptResponseCBC encrypts a response using Unity's CBC approach
// WARNING: This bypasses all security features (HMAC, timestamp, nonce)
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

	// Return as plain string (no SecurePackage wrapper)
	return []byte(encryptedData), nil
}
