package server

import (
	"crypto/rsa"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"log"
)

// EventType specifies what event occured
type EventType int

const (
	// ChatMessageEvent is an event describing a received chat message
	ChatMessageEvent EventType = iota
	// UserJoinedEvent is an event describing a user joining a chatroom
	UserJoinedEvent
)

// Event represents an event which will come over an internal channel
type Event struct {
	Type    EventType
	Message interface{}
}

// SendChatMessage is an event which encodes a user sending a message to
// others
type SendChatMessage struct {
	Sender           string
	Timestamp        int64
	EncryptedContent map[string][]byte
}

// ChatRoom contains the information about the clients in a chatroom and has
// the responsibility to distribute messages to the corrent correspondent
type ChatRoom struct {
	// All subscribers indexed based on username
	subscribers map[string]struct {
		sink   chan<- Event
		pubKey *rsa.PublicKey
	}
	// The channel where users can send a ChatClient to register for messages
	register chan struct {
		username string
		sink     chan<- Event
		pubKey   *rsa.PublicKey
	}
	// The publisher where any user can send a message to the chatroom
	broadcast chan SendChatMessage
}

// Run performs the event handling loop of the chatroom
func (room *ChatRoom) Run() {
	for {
		select {
		// Received a message which should be broadcasted to all subscribers
		case event := <-room.broadcast:

			// Template for a message
			sndmsg := websock.ChatMessage{
				Sender:    event.Sender,
				Timestamp: event.Timestamp,
			}

			// Iterate over all subscribers
			for recipent, info := range room.subscribers {

				// Get the encrypted content for the current recipent
				content, ok := event.EncryptedContent[recipent]
				if !ok {
					// TODO handle that msg did not include a cipher-message
					// for recipent
					continue
				}

				sndmsg.Message = content

				select {
				// Try to send message to clients sink
				case info.sink <- Event{ChatMessageEvent, sndmsg}:
				// TODO perhaps change, will currently close connection to all
				// clients which do not have room for another event
				default:
					close(info.sink)
					delete(room.subscribers, recipent)
				}
			}
		// Received a registration request
		case reg := <-room.register:
			if _, ok := room.subscribers[reg.username]; ok {
				log.Printf("User (%s) tried to subscribe to a chatroom multiple times\n", reg.username)
				continue
			}
			room.subscribers[reg.username] = struct {
				sink   chan<- Event
				pubKey *rsa.PublicKey
			}{
				reg.sink,
				reg.pubKey,
			}

			// TODO warn all existing users that a new user has joined
		}
	}
}

// Register registers a user to receive events from the chatroom
func (room *ChatRoom) Register(username string, pubKey *rsa.PublicKey, sink chan<- Event) {
	room.register <- struct {
		username string
		sink     chan<- Event
		pubKey   *rsa.PublicKey
	}{username, sink, pubKey}
}

// Broadcast sends an event to all subscribers
func (room *ChatRoom) Broadcast(event SendChatMessage) {
	room.broadcast <- event
}
