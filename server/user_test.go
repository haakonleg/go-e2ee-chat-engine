package server

import (
	"crypto/rsa"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
	"net/http/httptest"
	"testing"
)

var testserver *Server
var wsserver *httptest.Server
var priKey *rsa.PrivateKey
var pubKey *rsa.PublicKey

func init() {
	priKey, pubKey = setupTestKeys(2048)
	// Start server
	testserver, wsserver = setupTestServer()
}

func TestWebsocketConnetion(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}
}

func TestCreateUser(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	// Send a request to register the user
	err = websock.Send(ws, &websock.Message{
		Type: websock.RegisterUser,
		Message: &websock.RegisterUserMessage{
			Username:  "createuser",
			PublicKey: util.MarshalPublic(pubKey),
		},
	})
	if err != nil {
		t.Fatalf("Unable to send register user request: %s", err.Error())
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of register user was an error: %s\n", msg.Message.(string))
	default:
		t.Fatalf("Response of register user was non-ok (%d) type\n", msg.Type)

	}
}

func TestLoginUser(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	// Send a request to register the user
	err = websock.Send(ws, &websock.Message{
		Type: websock.RegisterUser,
		Message: &websock.RegisterUserMessage{
			Username:  "loginuser",
			PublicKey: util.MarshalPublic(pubKey),
		},
	})

	if err != nil {
		t.Fatalf("Unable to send register user request: %s", err.Error())
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of register user was an error: %s\n", msg.Message.(string))
	default:
		t.Fatalf("Response of register user was non-ok (%d) type\n", msg.Type)
	}

	// Send login request to server
	err = websock.Send(ws, &websock.Message{
		Type:    websock.LoginUser,
		Message: "loginuser",
	})

	// Receive auth challenge from server
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}
	switch msg.Type {
	case websock.AuthChallenge:
	case websock.Error:
		t.Fatalf("Response of login user was an error: %s\n", msg.Message.(string))
	default:
		t.Fatalf("Response of login user was non-auth-challenge (%d) type\n", msg.Type)
	}

	// Try to decrypt auth challenge
	decKey, err := rsa.DecryptPKCS1v15(nil, priKey, msg.Message.([]byte))
	if err != nil {
		t.Fatalf("Unable to decrypt auth challange: %s\n", err)
	}

	// Send decrypted auth key to server
	err = websock.Send(ws, &websock.Message{
		Type:    websock.AuthChallengeResponse,
		Message: decKey,
	})
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}

	// Receive auth challenge response from server
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of auth challenge response was an error: %s\n", msg.Message.(string))
	default:
		t.Fatalf("Response of auth challenge response was non-ok (%d) type\n", msg.Type)
	}
}

func TestInvalidUserData(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	validKey := util.MarshalPublic(pubKey)

	_, _smallkey := setupTestKeys(512)
	smallkey := util.MarshalPublic(_smallkey)

	_, _bigkey := setupTestKeys(4096)
	bigkey := util.MarshalPublic(_bigkey)

	for _, v := range []struct {
		name string
		key  []byte
	}{
		{"", validKey},
		{"a", validKey},
		{"aa", validKey},
		{"abcdefghijklmnopqrstuvxyz", validKey},
		{"__asda__", validKey},
		{"bbddl??", validKey},
		{"bbddl??\\", validKey},
		{"<script>alert(1)</script>", validKey},
		// Caused a panic in the server (fixed by #15)
		{"john1", []byte{0, 0, 0, 0, 1, 1, 1}},
		{"john2", nil},
		{"bigjohn", bigkey},
		{"smalljohn", smallkey},
	} {
		// Send a request to register the user
		err = websock.Send(ws, &websock.Message{
			Type: websock.RegisterUser,
			Message: &websock.RegisterUserMessage{
				Username:  v.name,
				PublicKey: v.key,
			},
		})

		if err != nil {
			t.Fatalf("Unable to send register user request: %s", err.Error())
		}

		msg := new(websock.Message)
		err = websock.Receive(ws, msg)
		if err != nil {
			t.Fatalf("Error when receiving message from server: %s\n", err)
		}
		switch msg.Type {
		case websock.OK:
			t.Fatalf("Response of register user (%s) was ok despite invalid\n", v.name)
		case websock.Error:
			t.Logf("Response was error as expected for user (%s): %s\n", v.name, msg.Message.(string))
		default:
			t.Fatalf("Response of register user was non-ok (%d) type for user (%s)\n", msg.Type, v.name)
		}
	}
}
