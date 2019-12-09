package crypto

import (
	"crypto/aes"
	"crypto/cipher"
)

// AESCBCEncrypt is given key, iv to encrypt the plainText in AES CBC way.
func AESCBCEncrypt(key, iv, plainText []byte) ([]byte, error) {

	if len(plainText)%aes.BlockSize != 0 {
		panic("plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plainText))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plainText)

	return ciphertext, nil

}

// AESCBCDecrypt is given key, iv to decrypt the cipherText in AES CBC way.
func AESCBCDecrypt(key, iv, cipherText []byte) ([]byte, error) {

	if len(cipherText) < aes.BlockSize {
		panic("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	cipherText = cipherText[aes.BlockSize:]

	if len(cipherText)%aes.BlockSize != 0 {
		panic("cipherText is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)
	return cipherText, nil
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
