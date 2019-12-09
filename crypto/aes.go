package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	padding := origData[len(origData)-1]
	return origData[:len(origData)-int(padding)]
}

// AESCBCEncrypt is given key, iv to encrypt the plainText in AES CBC way.
func AESCBCEncrypt(key, iv, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)

	content := pkcs5Padding(plainText, block.BlockSize())
	ciphertext := make([]byte, len(content))
	mode.CryptBlocks(ciphertext, content)

	return ciphertext, nil

}

// AESCBCDecrypt is given key, iv to decrypt the cipherText in AES CBC way.
func AESCBCDecrypt(key, iv, cipherText []byte) ([]byte, error) {

	if len(cipherText) == 0 {
		panic("ciphertext can't be plain")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	mode.CryptBlocks(plainText, cipherText)

	return pkcs5UnPadding(plainText), nil
}

// AESECBEncrypt is given key, iv to encrypt the plainText in AES ECB way.
func AESECBEncrypt(key, plainText []byte) ([]byte, error) {
	if len(plainText)%aes.BlockSize != 0 {
		panic("Need a multiple of the blocksize")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, len(plainText))

	for len(plainText) > 0 {
		block.Encrypt(cipherText, plainText[:aes.BlockSize])
		plainText = plainText[aes.BlockSize:]
		cipherText = cipherText[aes.BlockSize:]
	}
	return cipherText, nil
}

// AESECBDecrypt is given key, iv to decrypt the cipherText in AES ECB way.
func AESECBDecrypt(key, cipherText []byte) ([]byte, error) {
	if len(cipherText)%aes.BlockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plainText := make([]byte, len(cipherText))
	for len(cipherText) > 0 {
		block.Decrypt(plainText, cipherText[:aes.BlockSize])
		plainText = plainText[aes.BlockSize:]
		cipherText = cipherText[aes.BlockSize:]
	}

	return plainText, nil
}
