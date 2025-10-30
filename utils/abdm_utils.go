package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

func EncryptWithPublicKey(publicKeyPEM string, plainText string) (string, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode([]byte("-----BEGIN PUBLIC KEY-----\n" + publicKeyPEM + "\n-----END PUBLIC KEY-----"))
	if block == nil {
		return "", errors.New("failed to decode public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not RSA public key")
	}

	// Encrypt using RSA-OAEP with SHA-1 (as required by ABDM)
	cipherText, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, []byte(plainText), nil)
	if err != nil {
		return "", err
	}

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(cipherText), nil
}
