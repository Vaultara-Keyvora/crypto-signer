package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

// AESEncryptWithKey aes-128-cbc
func AESEncryptWithKey(plaintext, key string) (EncodeData string, err error) {

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return EncodeData, err
	}

	plaintextBytes := []byte(plaintext)
	paddedText := padPKCS7(plaintextBytes, aes.BlockSize)
	ciphertext := make([]byte, len(paddedText))
	mode := cipher.NewCBCEncrypter(block, []byte(key))
	mode.CryptBlocks(ciphertext, paddedText)
	return "v1:" + strings.ToUpper(hex.EncodeToString(ciphertext)), nil
}

// AESDecryptWithKey aes-128-cbc
func AESDecryptWithKey(ciphertext, key string) (DecryptData string, err error) {
	if !strings.HasPrefix(ciphertext, "v1:") {
		return ciphertext, nil
	}
	ciphertext = strings.TrimPrefix(ciphertext, "v1:")
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return DecryptData, err
	}
	deCt, err := hex.DecodeString(ciphertext)
	if err != nil {
		return DecryptData, err
	}
	pt := make([]byte, len(deCt))
	// CBC
	blockMode := cipher.NewCBCDecrypter(block, []byte(key))
	blockMode.CryptBlocks(pt, deCt)

	UnPaddedText, err := UnPadPKCS7(pt)
	if err != nil {
		return "", err
	}
	return string(UnPaddedText), nil
}

// UnPadPKCS7  PKCS#7 padding
func UnPadPKCS7(data []byte) ([]byte, error) {
	padding := int(data[len(data)-1])
	if len(data) < padding {
		return nil, errors.New("invalid data")
	}
	return data[:len(data)-padding], nil
}

// PKCS#7 padding
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, len(data)+padding)
	copy(padText, data)
	for i := len(data); i < len(padText); i++ {
		padText[i] = byte(padding)
	}
	return padText
}

func RandString(length uint) string {
	b := make([]byte, length/2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
