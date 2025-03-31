package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

// DeriveKeyFromPassword generates a master key from a password and email
func DeriveKeyFromPassword(password, email string) ([]byte, error) {
	salt := normalizeEmail(email)
	masterKey := pbkdf2.Key(
		[]byte(password),
		[]byte(salt),
		100000, // Iterations
		32,     // Key length
		sha256.New,
	)
	return masterKey, nil
}

// GenerateSymmetricKey creates a new random symmetric key
func GenerateSymmetricKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptSymmetricKey encrypts a symmetric key with the master key
func EncryptSymmetricKey(symmetricKey, masterKey []byte) (string, error) {
	encrypted, err := encrypt(symmetricKey, masterKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptSymmetricKey decrypts a symmetric key with the master key
func DecryptSymmetricKey(encryptedKeyBase64 string, masterKey []byte) ([]byte, error) {
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyBase64)
	if err != nil {
		return nil, err
	}
	return decrypt(encryptedKey, masterKey)
}

// encrypt performs AES-GCM encryption
func encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt performs AES-GCM decryption
func decrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// normalizeEmail standardizes an email for use as a salt
func normalizeEmail(email string) string {
	// In a production environment, you might want to do more sophisticated normalization
	// For now, we'll just lowercase the email
	return fmt.Sprintf("pastepal:%s", email)
}
