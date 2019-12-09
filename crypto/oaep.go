package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"fmt"
)

func OAEPDecrypt(pub *rsa.PublicKey, pri *rsa.PrivateKey, cipherText []byte) ([]byte, error) {
	if len(cipherText) == 0 {
		return nil, errors.New("cipherText can not be empty string.")
	}
	buffer := bytes.Buffer{}
	for _, cipherTextBlock := range grouping(cipherText, len(pub.N.Bytes())) {
		plainText, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, pri, cipherTextBlock, make([]byte, 0))
		if err != nil {

			return nil, err
		}
		buffer.Write(plainText)
	}
	return buffer.Bytes(), nil
}

func grouping(src []byte, size int) [][]byte {
	var groups [][]byte
	fmt.Println(size)
	srcSize := len(src)
	if srcSize <= size {
		groups = append(groups, src)
	} else {
		for len(src) != 0 {
			if len(src) <= size {
				groups = append(groups, src)
				break
			} else {
				v := src[:size]
				groups = append(groups, v)
				src = src[size:]
			}
		}
	}

	return groups
}
