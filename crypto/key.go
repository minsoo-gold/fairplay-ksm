package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
)

func DecryptPriKey(prikey, passphrase []byte) (*rsa.PrivateKey, error) {
	priPem, _ := pem.Decode(prikey)
	if priPem.Type != "RSA PRIVATE KEY" {
		panic("Private Key is not RSA Private Key.")
	}

	var decryptedPriKeyByte []byte
	if len(passphrase) == 0 {
		decryptedPriKeyByte = priPem.Bytes
	} else {
		decrypted, err := x509.DecryptPEMBlock(priPem, passphrase)
		if err != nil {
			return nil, err
		}
		decryptedPriKeyByte = decrypted
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(decryptedPriKeyByte)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func ParseASk(ask string) ([]byte, error) {
	if ask == "" {
		ask = "d87ce7a26081de2e8eb8acef3a6dc179"
	}
	parsedASk, err := hex.DecodeString(ask)
	if err != nil {
		return nil, err
	}
	return parsedASk, nil
}

func ParsePublicCertification(pubCert []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubCert)
	if block == nil {
		panic("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic("failed to parse certificate: " + err.Error())
	}
	return cert.PublicKey.(*rsa.PublicKey), nil
}
