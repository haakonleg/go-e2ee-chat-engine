package main

import (
	"crypto/rand"
	"crypto/rsa"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"golang.org/x/net/websocket"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

// ChatSession contains the context and callback methods of a chat session
type ChatSession struct {
	DisconnectFunc func()
	OnChatInfo     func(error, *ChatSession, *websock.ChatInfoMessage)
	OnChatMessage  func(error, *ChatSession, *websock.ChatMessage)
	OnUserJoined   func(error, *ChatSession, *websock.User)
	OnUserLeft     func(*ChatSession, string)
	Reader         *WSReader
	Socket         *websocket.Conn
	PrivateKey     *rsa.PrivateKey
	AuthKey        []byte

	username string
	users    map[string]*websock.User
}

// StartChatSession runs in a separate goroutine and listens for new chat messages and users when a user is in a chat session
func (cs *ChatSession) StartChatSession() {
	cs.users = make(map[string]*websock.User, 0)

Loop:
	for {
		msg, err := cs.Reader.GetNext()
		if err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case websock.ChatInfo:
			chatInfo := msg.Message.(*websock.ChatInfoMessage)
			cs.username = chatInfo.MyUsername

			// Decrypt chat messages
			err = cs.DecryptChatMessages(chatInfo.Messages...)

			// Add users to the user list
			for i := range chatInfo.Users {
				cs.users[chatInfo.Users[i].Username] = &chatInfo.Users[i]
			}
			cs.OnChatInfo(err, cs, chatInfo)

		case websock.ChatMessageReceived:
			chatMessage := msg.Message.(*websock.ChatMessage)
			cs.DecryptChatMessages(chatMessage)
			cs.OnChatMessage(err, cs, chatMessage)

		case websock.UserJoined:
			user := msg.Message.(*websock.User)
			cs.users[user.Username] = user
			cs.OnUserJoined(err, cs, user)

		case websock.UserLeft:
			// Remove user from the user list
			username := msg.Message.(string)

			// If the user who left is me, quit the chat
			if username == cs.username {
				break Loop
			}

			delete(cs.users, username)
			cs.OnUserLeft(cs, username)
		}
	}

	cs.DisconnectFunc()
}

// DecryptChatMessages decrypts chat messages using an RSA private key
func (cs *ChatSession) DecryptChatMessages(chatMessages ...*websock.ChatMessage) error {
	for i := range chatMessages {
		decMsg, err := rsa.DecryptPKCS1v15(rand.Reader, cs.PrivateKey, chatMessages[i].Message)
		if err != nil {
			return err
		}
		chatMessages[i].Message = decMsg
	}

	return nil
}

// SendChatMessage sends a chat message in the chat room of the chat session
// The message is encrypted with every participants public key, and sent to the server
func (cs *ChatSession) SendChatMessage(message string) {
	req := &websock.SendChatMessage{
		EncryptedContent: make(map[string][]byte)}

	// For every user in the chat, encrypt the message with their public key
	for _, user := range cs.users {
		pubKey, err := util.UnmarshalPublic(user.PublicKey)
		if err != nil {
			log.Println(err)
			continue
		}
		encMsg, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(message))
		if err != nil {
			log.Println(err)
			continue
		}
		req.EncryptedContent[user.Username] = encMsg
	}

	websock.Msg.Send(cs.Socket, &websock.Message{Type: websock.SendChat, Message: req})
}

// LeaveChat is called when a user decides to leave a chat room. The client sends a message
// notifying the server that the client has left the chat room.
func (cs *ChatSession) LeaveChat() {
	websock.Msg.Send(cs.Socket, &websock.Message{Type: websock.LeaveChat})
}
