package crypto

import (
	"encoding/base64"
	"errors"
	"io"
)

// EncryptData encrypts data using the provided symmetric key
func EncryptData(data []byte, symmetricKey []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.New("no data to encrypt")
	}

	encrypted, err := encrypt(data, symmetricKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptData decrypts data using the provided symmetric key
func DecryptData(encryptedDataBase64 string, symmetricKey []byte) ([]byte, error) {
	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataBase64)
	if err != nil {
		return nil, err
	}

	return decrypt(encryptedData, symmetricKey)
}

// EncryptReader encrypts data from a reader using the provided symmetric key
func EncryptReader(reader io.Reader, symmetricKey []byte) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return EncryptData(data, symmetricKey)
}

// DecryptToWriter decrypts data and writes it to the provided writer
func DecryptToWriter(encryptedDataBase64 string, symmetricKey []byte, writer io.Writer) error {
	data, err := DecryptData(encryptedDataBase64, symmetricKey)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}
