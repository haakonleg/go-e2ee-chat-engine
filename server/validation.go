package server

import (
	"strings"
	"unicode"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// Checks that a string only contains alphanumeric characters
func isAlphaNumeric(input string) bool {
	for _, ch := range input {
		if !unicode.IsLetter(ch) &&
			!unicode.IsNumber(ch) {
			return false
		}
	}
	return true
}

// ValidateRegisterUser validates the contents of a request from a client to
// register a new user. The length of the username and the public key bit-length is validated.
func ValidateRegisterUser(ws *websocket.Conn, msg *websock.RegisterUserMessage) bool {
	msg.Username = strings.TrimSpace(msg.Username)

	if len(msg.Username) < 3 {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Username must contain at least 3 characters"})
		return false
	} else if len(msg.Username) > 20 {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Username cannot contain more than 20 characters"})
		return false
	} else if !isAlphaNumeric(msg.Username) {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Username can only contain alphanumeric characters"})
		return false
	}

	// Check key length
	if pubKey, err := util.UnmarshalPublic(msg.PublicKey); err != nil {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Invalid public key"})
		return false
	} else if pubKey.N.BitLen() != 2048 {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Size of public key modulus is not 2048 bits"})
		return false
	}

	return true
}

// ValidateCreateChatRoom validates the content of a request from a client to create a new chat room.
// the name of the chat room is validated. If the chat room has a password, this is also validated.
func ValidateCreateChatRoom(ws *websocket.Conn, msg *websock.CreateChatRoomMessage) bool {
	if len(msg.Name) < 3 {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Chat room name must contain at least 3 characters"})
		return false
	} else if len(msg.Name) > 30 {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Chat room name cannot contain more than 30 characters"})
		return false
	} else if !isAlphaNumeric(msg.Name) {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Chat room name can only contain alphanumeric characters"})
		return false
	}

	if len(msg.Password) != 0 {
		if len(msg.Password) < 6 {
			websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Password must contain at least 6 characters"})
			return false
		} else if len(msg.Password) > 60 {
			websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Password cannot contain more than 60 characters"})
			return false
		}
	}
	return true
}
