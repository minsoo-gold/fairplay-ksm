package cryptos

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/youmark/pkcs8"
)

/*
func DecryptPriKey(prikey, passphrase []byte) (*rsa.PrivateKey, error) {
	priPem, _ := pem.Decode(prikey)
	if priPem.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("private key is not RSA Private Key")
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
}*/

func DecryptPriKey(prikey, passphrase []byte) (*rsa.PrivateKey, error) {
	priPem, _ := pem.Decode(prikey)
	if priPem == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch priPem.Type {
	case "RSA PRIVATE KEY": // PKCS#1
		var der []byte
		var err error
		if x509.IsEncryptedPEMBlock(priPem) {
			if len(passphrase) == 0 {
				return nil, fmt.Errorf("private key is encrypted but no passphrase provided")
			}
			der, err = x509.DecryptPEMBlock(priPem, passphrase)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt PKCS#1: %w", err)
			}
		} else {
			der = priPem.Bytes
		}
		return x509.ParsePKCS1PrivateKey(der)

	case "ENCRYPTED PRIVATE KEY": // PKCS#8 (encrypted)
		if len(passphrase) == 0 {
			return nil, fmt.Errorf("passphrase required for ENCRYPTED PRIVATE KEY")
		}
		anyKey, err := pkcs8.ParsePKCS8PrivateKey(priPem.Bytes, passphrase)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt PKCS#8: %w", err)
		}
		rsaKey, ok := anyKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("decrypted PKCS#8 is not RSA")
		}
		return rsaKey, nil

	case "PRIVATE KEY": // PKCS#8 (unencrypted)
		anyKey, err := x509.ParsePKCS8PrivateKey(priPem.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS#8: %w", err)
		}
		rsaKey, ok := anyKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("PRIVATE KEY is not RSA")
		}
		return rsaKey, nil
	}

	return nil, fmt.Errorf("unsupported private key type: %s", priPem.Type)
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

func ParsePublicCertification(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse certificate: %w", err)
		}
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate public key is not RSA")
		}
		return pub, nil

	case "CERTIFICATE REQUEST": // a.k.a. "NEW CERTIFICATE REQUEST"
		req, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse CSR: %w", err)
		}
		// CSR 자체 서명 검증(선택)
		if err := req.CheckSignature(); err != nil {
			return nil, fmt.Errorf("CSR signature invalid: %w", err)
		}
		pub, ok := req.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("CSR public key is not RSA")
		}
		return pub, nil

	default:
		return nil, fmt.Errorf("unsupported PEM type: %s", block.Type)
	}
}
