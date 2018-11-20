package server

import (
	"crypto/rsa"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
	"testing"
)

func TestCreateValidUsers(t *testing.T) {
	pkm := util.MarshalPublic(pubkey)
	for _, username := range []string{
		"john",
		"chris",
		"mandananla2",
	} {
		ws, err := websocket.Dial(wsserver.URL, "", "http://")
		defer ws.Close()
		if err != nil {
			t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
		}
		if err := registerUser(ws, username, pkm); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoginValidUser(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	if err := registerUser(ws, "validuser", util.MarshalPublic(pubkey)); err != nil {
		t.Fatal(err)
	}

	if err := loginUser(ws, "validuser", prikey); err != nil {
		t.Fatal(err)
	}

}

func TestLoginNonexistentUsername(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	// Send login request to server
	err = websock.Send(ws, &websock.Message{
		Type:    websock.LoginUser,
		Message: "doesnotexist",
	})
	if err != nil {
		t.Fatalf("Unable to send message to server: %s\n", err)
	}

	// Receive auth challenge from server
	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s\n", err)
	}

	switch msg.Type {
	case websock.Error:
	case websock.AuthChallenge:
		t.Fatalf("Response of login user for non-existent user was an auth challenge")
	default:
		t.Fatalf("Response of login user was non-error type (%d)", msg.Type)
	}
}

func TestLoginInvalidKeyUser(t *testing.T) {
	ws, err := websocket.Dial(wsserver.URL, "", "http://")
	defer ws.Close()
	if err != nil {
		t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
	}

	if err := registerUser(ws, "invalidkeyuser", util.MarshalPublic(pubkey)); err != nil {
		t.Fatal(err)
	}

	// Send login request to server
	err = websock.Send(ws, &websock.Message{
		Type:    websock.LoginUser,
		Message: "invalidkeyuser",
	})
	if err != nil {
		t.Fatalf("Unable to send message to server: %s\n", err)
	}

	// Receive auth challenge from server
	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Error when receiving message from server: %s", err)
	}
	switch msg.Type {
	case websock.AuthChallenge:
	case websock.Error:
		t.Fatalf("Response of login user was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of login user was non-auth-challenge type (%d)", msg.Type)
	}

	// Try to decrypt auth challenge
	_, err = rsa.DecryptPKCS1v15(nil, sprikey, msg.Message.([]byte))
	if err == nil {
		t.Fatal("Was able to decrypt auth challenge with wrong private key")
	}
}

func TestRegisterInvalidUser(t *testing.T) {

	validKey := util.MarshalPublic(pubkey)
	smallkey := util.MarshalPublic(invalidsmallpubkey)
	bigkey := util.MarshalPublic(invalidbigpubkey)

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

		ws, err := websocket.Dial(wsserver.URL, "", "http://")
		defer ws.Close()
		if err != nil {
			t.Fatalf("Unable to connect to websocket at '%s': %s\n", wsserver.URL, err)
		}

		if err := registerUser(ws, v.name, v.key); err != nil {
			t.Logf("Got expected error for (%s): %s", v.name, err)
		} else {
			t.Fatalf("Got unexpected ok when registration (%s)", v.name)
		}
	}
}
