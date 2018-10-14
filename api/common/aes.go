package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

//Encrypt 加密字符串
func Encrypt(str string) (string, error) {
	encryptData, err := AesEncrypt([]byte(str), Aeskey)
	if err != nil {
		return "", err
	}
	myBase64 := base64.StdEncoding
	content := make([]byte, myBase64.EncodedLen(len(encryptData)))
	myBase64.Encode(content, encryptData)
	return string(content), nil
}

//Decrypt 解密字符串
func Decrypt(str string) (string, error) {
	myBase64 := base64.StdEncoding
	encryptData, err := myBase64.DecodeString(str)
	if err != nil {
		return "", err
	}
	var dencryptData []byte
	dencryptData, err = AesEncrypt(encryptData, Aeskey)
	if err != nil {
		return "", err
	}
	return string(dencryptData), nil
}

//AesEncrypt AES加密
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs7Encoder(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AesDecrypt AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs7Decoder(origData)
	return origData, nil
}

func pkcs7Encoder(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs7Decoder(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
