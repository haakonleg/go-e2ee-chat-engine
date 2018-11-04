package user

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	mrand "math/rand"
)

const (
	bitSize     = 2048
	authKeyLen  = 64
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// GenKeyPair generates an RSA key pair
func GenKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
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
func UnmarshalPrivate(pemBlock []byte) (*rsa.PrivateKey, error) {
	data, _ := pem.Decode(pemBlock)
	key, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// MarshalPublic marshals an RSA public key to a byte slice
func MarshalPublic(key *rsa.PublicKey) []byte {
	pemBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key)}

	return pem.EncodeToMemory(pemBlock)
}

// UnmarshalPublic unmarshals an RSA public key from byte format
func UnmarshalPublic(pemBlock []byte) (*rsa.PublicKey, error) {
	data, _ := pem.Decode(pemBlock)
	key, err := x509.ParsePKCS1PublicKey(data.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// Generates a random authentication key, and encrypts it with the given public key
// returns the encrypted and the original auth key
func genAuthChallenge(pubKey *rsa.PublicKey) ([]byte, []byte) {
	authKey := make([]byte, authKeyLen)
	for i := range authKey {
		authKey[i] = letterBytes[mrand.Int63()%int64(len(letterBytes))]
	}

	encKey, _ := rsa.EncryptPKCS1v15(rand.Reader, pubKey, authKey)
	return encKey, authKey
}
