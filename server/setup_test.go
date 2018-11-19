package server

import (
	"crypto/rand"
	"crypto/rsa"
	"golang.org/x/net/websocket"
	"log"
	"net/http/httptest"
	"os"
	"strings"
)

// setupTestServer creates a test server using a httptest.Server
//
// WARNING: This function will flush the provided database to provide a clean
// database for insertions
func setupTestServer() (testserver *Server, wsserver *httptest.Server) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("Error: environment variable MONGODB_URI is not set")
	}
	dbName := os.Getenv("MONGODB_NAME")
	if dbName == "" {
		log.Fatal("Error: environment variable MONGODB_NAME is not set")
	}

	serverConfig := Config{
		DBName:    dbName,
		MongoURL:  mongoURI,
		Keepalive: 100,
	}

	testserver = CreateServer(serverConfig)

	// Flush database
	testserver.Db.DeleteAll()

	wsserver = httptest.NewServer(websocket.Handler(testserver.WebsockHandler))

	// Change protocol from http to ws
	wsserver.URL = "ws" + strings.TrimPrefix(wsserver.URL, "http")

	return
}

// setupTestKeys creates a dummy private and a public rsa keypair
func setupTestKeys(testKeySize int) (priKey *rsa.PrivateKey, pubKey *rsa.PublicKey) {
	var err error

	if priKey, err = rsa.GenerateKey(rand.Reader, testKeySize); err != nil {
		log.Fatalf("Unable to generate private key: %s\n", err)
	}
	pubKey = &priKey.PublicKey

	return
}
