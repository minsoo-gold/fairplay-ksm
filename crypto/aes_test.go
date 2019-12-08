package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCBCEncrypt(t *testing.T) {
	assert := assert.New(t)
	var key = "1ae8ccd0e7985cc0b6203a55855a1034afc252980e970ca90e5202689f947ab9"
	var iv = "d58ce954203b7c9a9a9d467f59839249"

	keyByteAry, _ := hex.DecodeString(key)
	ivByteAry, _ := hex.DecodeString(iv)
	plainText := []byte("6368616e676520746869732070617373")

	crypted, err := AESCBCEncrypt(keyByteAry, ivByteAry, plainText)

	assert.NoError(err)
	assert.Equal("AAAAAAAAAAAAAAAAAAAAAH63109cCQq85FYwIRciIck+HN+B5yJyFueE8ZN1zLI4", base64.StdEncoding.EncodeToString(crypted))

}

func TestCBCDecrypt(t *testing.T) {
	assert := assert.New(t)
	var key = "1ae8ccd0e7985cc0b6203a55855a1034afc252980e970ca90e5202689f947ab9"
	var iv = "d58ce954203b7c9a9a9d467f59839249"

	keyByteAry, _ := hex.DecodeString(key)
	ivByteAry, _ := hex.DecodeString(iv)

	enBase64Str := "AAAAAAAAAAAAAAAAAAAAAH63109cCQq85FYwIRciIck+HN+B5yJyFueE8ZN1zLI4"

	en, err := base64.StdEncoding.DecodeString(enBase64Str)
	assert.NoError(err)

	plainText, err := AESCBCDecrypt(keyByteAry, ivByteAry, en)

	assert.NoError(err)
	assert.Equal("6368616e676520746869732070617373", string(plainText))
}
