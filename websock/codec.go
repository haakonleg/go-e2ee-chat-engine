package websock

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"

	"golang.org/x/net/websocket"
)

// codec is the decoder for messages sent over the websocket
var codec = websocket.Codec{Marshal: marshalMessage, Unmarshal: unmarshalMessage}

// Send sends a message to the connection synchronously
//
// NB! This will fail if the given message contains incompatible type and content
func Send(ws *websocket.Conn, msg *Message) error {
	return codec.Send(ws, msg)
}

// Receive fetches a message from the connection synchronously
//
// NB! This will fail if the received message contains incompatible type and content
func Receive(ws *websocket.Conn, msg *Message) error {
	return codec.Receive(ws, msg)
}

// Register types for gob encoding/decoding
func init() {
	gob.Register(&RegisterUserMessage{})
	gob.Register(&CreateChatRoomMessage{})
	gob.Register(&GetChatRoomsResponseMessage{})
	gob.Register(&JoinChatMessage{})
	gob.Register(&ChatInfoMessage{})
	gob.Register(&ChatMessage{})
	gob.Register(&SendChatMessage{})
}

func marshalMessage(v interface{}) ([]byte, byte, error) {
	msg, ok := v.(*Message)
	if !ok {
		return nil, websocket.TextFrame, errors.New("Input to marshalMessage was not of type *Message")
	}

	if err := checkType(msg.Message, msg.Type); err != nil {
		return nil, websocket.TextFrame, err
	}

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(msg); err != nil {
		return nil, websocket.TextFrame, err
	}

	return buf.Bytes(), websocket.TextFrame, nil
}

func unmarshalMessage(data []byte, payloadType byte, v interface{}) error {
	msg, ok := v.(*Message)
	if !ok {
		return errors.New("Input to unmarshalMessage was not of type *Message")
	}

	reader := bytes.NewReader(data)
	dec := gob.NewDecoder(reader)
	if err := dec.Decode(msg); err != nil {
		return err
	}

	if err := checkType(msg.Message, msg.Type); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func checkType(v interface{}, msgType MessageType) error {
	switch msgType {
	case Error, OK, LoginUser, UserLeft:
		if _, ok := v.(string); !ok {
			return errors.New("Expected message type string")
		}

	case RegisterUser:
		if _, ok := v.(*RegisterUserMessage); !ok {
			return errors.New("Expected message type *RegisterUserMessage")
		}

	case AuthChallenge, AuthChallengeResponse:
		if _, ok := v.([]byte); !ok {
			return errors.New("Expected message type []byte")
		}

	case CreateChatRoom:
		if _, ok := v.(*CreateChatRoomMessage); !ok {
			return errors.New("Expected message type *CreateChatRoomMessage")
		}

	case GetChatRooms, LeaveChat, Ping, Pong:
		if v != nil {
			return errors.New("Expected message to be nil")
		}

	case GetChatRoomsResponse:
		if _, ok := v.(*GetChatRoomsResponseMessage); !ok {
			return errors.New("Expected message type *GetChatRoomsResponseMessage")
		}

	case JoinChat:
		if _, ok := v.(*JoinChatMessage); !ok {
			return errors.New("Expected message type *JoinChatMessage")
		}

	case ChatInfo:
		if _, ok := v.(*ChatInfoMessage); !ok {
			return errors.New("Expected message type *ChatInfoMessage")
		}

	case SendChat:
		if _, ok := v.(*SendChatMessage); !ok {
			return errors.New("Expected message type *SendChatMessage")
		}

	case ChatMessageReceived:
		if _, ok := v.(*ChatMessage); !ok {
			return errors.New("Expected message type *ChatMessage")
		}

	case UserJoined:
		if _, ok := v.(*User); !ok {
			return errors.New("Expected message type *User")
		}
	default:
		return errors.New("Invalid message type")
	}

	return nil
}
