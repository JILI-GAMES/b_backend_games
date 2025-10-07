package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type SecurePackage struct {
	Data      string `json:"data"`
	Signature string `json:"signature"`
}

type SecureEnvelope struct {
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Payload   string `json:"payload"`
}

var (
	AESKey  []byte
	HMACKey []byte
	// RequestExpirationWindow is the time window in milliseconds for request expiration
	RequestExpirationWindow int64 = 30000 // 30 seconds default
)

// Initialize loads keys from environment variables
func Initialize() error {
	aesKeyB64 := os.Getenv("ENCRYPTION_AES_KEY")
	hmacKeyB64 := os.Getenv("ENCRYPTION_HMAC_KEY")
	expirationWindow := os.Getenv("ENCRYPTION_EXPIRATION_WINDOW")

	if aesKeyB64 == "" || hmacKeyB64 == "" {
		return errors.New("encryption keys not found in environment")
	}

	// Set custom expiration window if provided
	if expirationWindow != "" {
		if window, err := strconv.ParseInt(expirationWindow, 10, 64); err == nil {
			RequestExpirationWindow = window
		}
	}

	var err error
	AESKey, err = base64.StdEncoding.DecodeString(aesKeyB64)
	if err != nil {
		return err
	}

	HMACKey, err = base64.StdEncoding.DecodeString(hmacKeyB64)
	if err != nil {
		return err
	}

	if len(AESKey) != 32 || len(HMACKey) != 32 {
		return errors.New("keys must be 32 bytes")
	}

	log.Println("Encryption keys loaded successfully")
	return nil
}

func EncryptAESGCM(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func DecryptAESGCM(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func CreateHMAC(data []byte) string {
	h := hmac.New(sha256.New, HMACKey)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func VerifyHMAC(data []byte, signature string) bool {
	expected := CreateHMAC(data)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

func DecryptRequest(body []byte) ([]byte, string, error) {
	var pkg SecurePackage
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, "", err
	}

	encryptedData, err := base64.StdEncoding.DecodeString(pkg.Data)
	if err != nil {
		return nil, "", err
	}

	if !VerifyHMAC(encryptedData, pkg.Signature) {
		return nil, "", errors.New("HMAC verification failed")
	}

	decryptedData, err := DecryptAESGCM(encryptedData)
	if err != nil {
		return nil, "", err
	}

	var envelope SecureEnvelope
	if err := json.Unmarshal(decryptedData, &envelope); err != nil {
		return nil, "", err
	}

	if time.Now().UnixMilli()-envelope.Timestamp > RequestExpirationWindow {
		return nil, "", errors.New("request expired")
	}

	return []byte(envelope.Payload), envelope.Nonce, nil
}

func EncryptResponse(data interface{}, nonce string) ([]byte, error) {
	responseJSON, _ := json.Marshal(data)

	envelope := SecureEnvelope{
		Timestamp: time.Now().UnixMilli(),
		Nonce:     nonce,
		Payload:   string(responseJSON),
	}

	envelopeJSON, _ := json.Marshal(envelope)
	encryptedData, err := EncryptAESGCM(envelopeJSON)
	if err != nil {
		return nil, err
	}

	signature := CreateHMAC(encryptedData)

	pkg := SecurePackage{
		Data:      base64.StdEncoding.EncodeToString(encryptedData),
		Signature: signature,
	}

	return json.Marshal(pkg)
}
