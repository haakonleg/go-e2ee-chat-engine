package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"time"
)

// GenKeyPair generates an RSA key pair
func GenKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	return key, &key.PublicKey
}

// MarshalPrivate marshals an RSA private key to a byte slice
func MarshalPrivate(key *rsa.PrivateKey) []byte {
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(pemBlock)
}

// UnmarshalPrivate unmarshals an RSA private key from byte format
func UnmarshalPrivate(pemBlock []byte) (key *rsa.PrivateKey, err error) {
	data, _ := pem.Decode(pemBlock)
	if data == nil {
		err = errors.New("Private key was not in the correct PEM format")
		return
	}
	key, err = x509.ParsePKCS1PrivateKey(data.Bytes)
	return
}

// MarshalPublic marshals an RSA public key to a byte slice
func MarshalPublic(key *rsa.PublicKey) []byte {
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key)}

	return pem.EncodeToMemory(pemBlock)
}

// UnmarshalPublic unmarshals an RSA public key from byte format
func UnmarshalPublic(pemBlock []byte) (key *rsa.PublicKey, err error) {
	data, _ := pem.Decode(pemBlock)
	if data == nil {
		err = errors.New("Public key was not in the correct PEM format")
		return
	}
	key, err = x509.ParsePKCS1PublicKey(data.Bytes)
	return
}

// NowMillis returns the current unix millisecond timestamp
func NowMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
