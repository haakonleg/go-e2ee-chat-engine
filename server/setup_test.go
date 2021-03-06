package server

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

var testserver *Server
var wsserver *httptest.Server

// Valid keys
var prikey *rsa.PrivateKey
var pubkey *rsa.PublicKey
var sprikey *rsa.PrivateKey
var spubkey *rsa.PublicKey

// Invalid keys
var invalidsmallpubkey *rsa.PublicKey
var invalidsmallprikey *rsa.PrivateKey
var invalidbigpubkey *rsa.PublicKey
var invalidbigprikey *rsa.PrivateKey

func init() {
	prikey, pubkey = setupTestKeys(2048)
	sprikey, spubkey = setupTestKeys(2048)

	invalidsmallprikey, invalidsmallpubkey = setupTestKeys(512)
	invalidbigprikey, invalidbigpubkey = setupTestKeys(4096)

	// Start server
	testserver, wsserver = setupTestServer()
}

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
		Keepalive: 100000,
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

func setupTestUser(username string, pk *rsa.PublicKey, pki *rsa.PrivateKey) (ws *websocket.Conn, err error) {
	ws, err = websocket.Dial(wsserver.URL, "", "http://")
	if err != nil {
		err = fmt.Errorf("Unable to connect to websocket at '%s': %s", wsserver.URL, err)
		return
	}

	if err = registerUser(ws, username, util.MarshalPublic(pk)); err != nil {
		return
	}

	err = loginUser(ws, username, pki)

	return
}

func registerUser(ws *websocket.Conn, username string, pubkey []byte) error {
	// Send a request to register the user
	err := websock.Send(ws, &websock.Message{
		Type: websock.RegisterUser,
		Message: &websock.RegisterUserMessage{
			Username:  username,
			PublicKey: pubkey,
		},
	})
	if err != nil {
		return fmt.Errorf("Unable to send register user request: %s", err)
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		return fmt.Errorf("Error when receiving message from server: %s", err)
	}
	switch msg.Type {
	case websock.OK:
		return nil
	case websock.Error:
		return fmt.Errorf("Response of register user was an error: %s", msg.Message.(string))
	default:
		return fmt.Errorf("Response of register user was non-ok type (%d)", msg.Type)
	}
}

func loginUser(ws *websocket.Conn, username string, pki *rsa.PrivateKey) error {
	err := websock.Send(ws, &websock.Message{
		Type:    websock.LoginUser,
		Message: username,
	})
	if err != nil {
		return fmt.Errorf("Unable to send register user request: %s", err)
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		return fmt.Errorf("Error when receiving message from server: %s", err)
	}
	switch msg.Type {
	case websock.AuthChallenge:
	case websock.Error:
		return fmt.Errorf("Response of login user was an error: %s", msg.Message.(string))
	default:
		return fmt.Errorf("Response of login user was non-auth challenge type (%d)", msg.Type)
	}

	// Try to decrypt auth challenge
	decKey, err := rsa.DecryptPKCS1v15(nil, pki, msg.Message.([]byte))
	if err != nil {
		return fmt.Errorf("Unable to decrypt auth challenge: %s", err)
	}

	// Send decrypted auth key to server
	err = websock.Send(ws, &websock.Message{
		Type:    websock.AuthChallengeResponse,
		Message: decKey,
	})
	if err != nil {
		return fmt.Errorf("Error when receiving message from server: %s", err)
	}

	// Receive auth challenge response from server
	err = websock.Receive(ws, msg)
	if err != nil {
		return fmt.Errorf("Error when receiving message from server: %s", err)
	}
	switch msg.Type {
	case websock.OK:
		return nil
	case websock.Error:
		return fmt.Errorf("Response of auth challenge response was an error: %s", msg.Message.(string))
	default:
		return fmt.Errorf("Response of auth challenge response was non-ok type (%d)", msg.Type)
	}

}
