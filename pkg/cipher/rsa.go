package cipher

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// 生成RSA密钥对
func GenRSA() (pubKey string, priKey string, fingerprint string, err error) {
	// 生成RSA密钥对，指定密钥长度（单位：位）
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("生成私钥失败：", err)
		return "", "", "", err
	}

	// 获取公钥
	publicKey := &privateKey.PublicKey

	// 将私钥序列化为PEM格式
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// 将公钥序列化为PEM格式
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	})

	// 计算SHA256指纹
	hash := sha256.Sum256(x509.MarshalPKCS1PublicKey(publicKey))

	pubKey = base64.StdEncoding.EncodeToString(publicKeyPEM)
	priKey = base64.StdEncoding.EncodeToString(privateKeyPEM)
	fingerprint = hex.EncodeToString(hash[:])
	return pubKey, priKey, fingerprint, nil
}

// 公钥指纹
func GetPublicKeyFingerprint(privateKey *rsa.PrivateKey) (string, error) {
	// 序列化公钥
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	// 计算SHA256指纹
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[:]), nil
}

// 公钥加密
func RSAEncrypt(pubKey string, plainText string) (string, error) {
	pubKeybuf, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}
	//pem解码
	block, _ := pem.Decode(pubKeybuf)
	pubKeyBytes, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	//使用公钥加密
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKeyBytes, []byte(plainText))

	if err != nil {
		return "", err
	}
	cipherTextBase64 := base64.StdEncoding.EncodeToString(cipherText)

	return cipherTextBase64, nil
}

// 私钥解密
func RSADecrypt(privKey string, cipherText string) ([]byte, error) {
	privKeybuf, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	//pem解码
	block, _ := pem.Decode(privKeybuf)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//使用私钥解密
	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherTextBytes)

	if err != nil {
		return nil, err
	}

	return plainText, nil
}
