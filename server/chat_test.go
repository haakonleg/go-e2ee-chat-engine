package server

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

func TestCreateChatRoom(t *testing.T) {
	ws, err := setupTestUser("createroom", pubkey, prikey)
	if err != nil {
		t.Fatal(err)
	}

	err = websock.Send(ws, &websock.Message{
		Type: websock.CreateChatRoom,
		Message: &websock.CreateChatRoomMessage{
			Name:     "createroom",
			Password: "",
			IsHidden: false,
		},
	})
	if err != nil {
		t.Fatalf("Unable to send create room request: %s", err)
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive create room response: %s", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of create room was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of create room was non-ok type (%d)", msg.Type)
	}
}

func TestJoinChatRoom(t *testing.T) {
	ws, err := setupTestUser("joinroom", pubkey, prikey)
	if err != nil {
		t.Fatal(err)
	}

	err = websock.Send(ws, &websock.Message{
		Type: websock.CreateChatRoom,
		Message: &websock.CreateChatRoomMessage{
			Name:     "joinroom",
			Password: "",
			IsHidden: false,
		},
	})
	if err != nil {
		t.Fatalf("Unable to send create room request: %s", err)
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive create room response: %s", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of create room was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of create room was non-ok type (%d)", msg.Type)
	}

	err = websock.Send(ws, &websock.Message{
		Type: websock.JoinChat,
		Message: &websock.JoinChatMessage{
			Name:     "joinroom",
			Password: "",
		},
	})
	if err != nil {
		t.Fatalf("Unable to send join room request: %s", err)
	}

	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive join room response: %s", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of join room was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of join room was non-ok type (%d)", msg.Type)
	}
}

func TestSendChatMessage(t *testing.T) {
	ws, err := setupTestUser("sendmsg", pubkey, prikey)
	if err != nil {
		t.Fatal(err)
	}

	err = websock.Send(ws, &websock.Message{
		Type: websock.CreateChatRoom,
		Message: &websock.CreateChatRoomMessage{
			Name:     "sendmsg",
			Password: "",
			IsHidden: false,
		},
	})
	if err != nil {
		t.Fatalf("Unable to send create room request: %s", err)
	}

	msg := new(websock.Message)
	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive create room response: %s", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of create room was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of create room was non-ok type (%d)", msg.Type)
	}

	err = websock.Send(ws, &websock.Message{
		Type: websock.JoinChat,
		Message: &websock.JoinChatMessage{
			Name:     "sendmsg",
			Password: "",
		},
	})
	if err != nil {
		t.Fatalf("Unable to send join room request: %s", err)
	}

	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive join room response: %s", err)
	}
	switch msg.Type {
	case websock.OK:
	case websock.Error:
		t.Fatalf("Response of join room was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response of join room was non-ok type (%d)", msg.Type)
	}

	err = websock.Receive(ws, msg)
	if err != nil {
		t.Fatalf("Unable to receive chat info response: %s", err)
	}
	switch msg.Type {
	case websock.ChatInfo:
	case websock.Error:
		t.Fatalf("Response was an error: %s", msg.Message.(string))
	default:
		t.Fatalf("Response was non-chat info type (%d)", msg.Type)
	}

	chatInfo := msg.Message.(*websock.ChatInfoMessage)

	req := &websock.SendChatMessage{
		EncryptedContent: make(map[string][]byte)}

	for _, user := range chatInfo.Users {
		// For every user in the chat, encrypt the message with their public key
		pubKey, err := util.UnmarshalPublic(user.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		encMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte("aaa"))
		if err != nil {
			t.Fatal(err)
		}
		req.EncryptedContent[user.Username] = encMsg
	}

	err = websock.Send(ws, &websock.Message{
		Type:    websock.SendChat,
		Message: req,
	})
	if err != nil {
		t.Fatalf("Unable to send chat message request: %s", err)
	}
}
