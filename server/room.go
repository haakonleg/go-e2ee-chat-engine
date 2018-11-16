package server

import (
	"crypto/rsa"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"log"
)

const subscriberSinkSize = 3
const registerSize = 5
const unregisterSize = 5
const broadcastSize = 5
const publisherSize = 10

// ChatRoom contains the information about the clients in a chatroom and has
// the responsibility to distribute messages to the corrent correspondent
type ChatRoom struct {
	// The internal representation of this chatroom in the database
	*mdb.Chat
	// All subscribers indexed based on username
	subscribers map[string]struct {
		sink   chan<- websock.Message
		pubKey *rsa.PublicKey
	}
	// The channel where users can send a ChatClient to register for messages
	register chan struct {
		username string
		sink     chan<- websock.Message
		pubKey   *rsa.PublicKey
	}
	// The channel where users can unsubscribe from the chatroom
	unregister chan string
	// The channel where incoming messages are broadcasted to all subscribers
	broadcast chan websock.Message
	// The channel where messages have a destination subscriber
	publisher chan struct {
		username string
		msg      websock.Message
	}
	stop chan struct{}
}

// NewChatRoom returns a new chatroom with the given name
func NewChatRoom(name, password string, isHidden bool) *ChatRoom {
	chatdb := mdb.NewChat(name, password, isHidden)
	return &ChatRoom{
		chatdb,
		make(map[string]struct {
			sink   chan<- websock.Message
			pubKey *rsa.PublicKey
		}),
		make(chan struct {
			username string
			sink     chan<- websock.Message
			pubKey   *rsa.PublicKey
		}, registerSize),
		make(chan string, unregisterSize),
		make(chan websock.Message, broadcastSize),
		make(chan struct {
			username string
			msg      websock.Message
		}, publisherSize),
		make(chan struct{}),
	}
}

// Run performs the event handling loop of the chatroom
func (room *ChatRoom) Run() {
	for {
		select {
		// Send a message to all subscribers
		case msg := <-room.broadcast:
			for recipentname, recipent := range room.subscribers {
				room.trySendMsg(recipentname, recipent.sink, msg)
			}
		// Send a message to a specific subscriber
		case msg := <-room.publisher:
			if recipent, ok := room.subscribers[msg.username]; ok {
				room.trySendMsg(msg.username, recipent.sink, msg.msg)
			}
		// Register a new subscriber
		case reg := <-room.register:
			room.registerSubscriber(reg.username, reg.pubKey, reg.sink)
		// Unregister a subscriber
		case username := <-room.unregister:
			delete(room.subscribers, username)

			// Send user left message to all subscribers
			room.Broadcast(websock.Message{
				Type:    websock.UserLeft,
				Message: username,
			})
		case <-room.stop:
			log.Printf("Shutting down chat ('%s')\n", room.Name)
			// TODO store state to mongo
			return
		}
	}
}

// Stop the chatroom eventloop
func (room *ChatRoom) Stop() {
	room.stop <- struct{}{}
}

// Subscribe registers a user to receive events from the chatroom
func (room *ChatRoom) Subscribe(username string, pubKey *rsa.PublicKey) <-chan websock.Message {
	sink := make(chan websock.Message, subscriberSinkSize)
	room.register <- struct {
		username string
		sink     chan<- websock.Message
		pubKey   *rsa.PublicKey
	}{username, sink, pubKey}
	return sink
}

// Unsubscribe unregisters a user from the chatroom
func (room *ChatRoom) Unsubscribe(username string) {
	room.unregister <- username
}

// Broadcast sends a message to all subscribers
func (room *ChatRoom) Broadcast(msg websock.Message) {
	room.broadcast <- msg
}

// Publish sends a message to specific subscriber if the subscriber exists
func (room *ChatRoom) Publish(username string, msg websock.Message) {
	room.publisher <- struct {
		username string
		msg      websock.Message
	}{username, msg}
}

// registerSubscriber registers a subscriber to the chatroom
func (room *ChatRoom) registerSubscriber(username string, pubKey *rsa.PublicKey, sink chan<- websock.Message) {

	if _, ok := room.subscribers[username]; ok {
		log.Printf("User (%s) tried to subscribe to a chatroom multiple times\n", username)
		sink <- websock.Message{
			Type:    websock.Error,
			Message: "Username is already a part of the chat"}

		close(sink)
		return
	}

	websockuser := websock.User{
		Username:  username,
		PublicKey: util.MarshalPublic(pubKey),
	}

	// Warn all current subscribers that a user joined
	evt := websock.Message{
		Type:    websock.UserJoined,
		Message: websockuser,
	}

	chatInfo := &websock.ChatInfoMessage{
		MyUsername: username,
		Users: []websock.User{
			websockuser,
		},
		Messages: make([]*websock.ChatMessage, 0)}

	// Iterate over all subscribers, send message and append to chatusers if
	// successfully sent message
	for otherusername, otheruser := range room.subscribers {
		if room.trySendMsg(otherusername, otheruser.sink, evt) {
			chatInfo.Users = append(chatInfo.Users, websock.User{
				Username:  otherusername,
				PublicKey: util.MarshalPublic(otheruser.pubKey)})
		}
	}

	// TODO find a way to access the messages for this user

	// Add the chat messages addressed to this user
	//for _, message := range s.FindMessagesForUser(user.Username, chatName) {
	//	// Check if the message actually has the encrypted message
	//	if len(message.MessageContent) == 0 {
	//		continue
	//	}

	//	chatInfo.Messages = append(chatInfo.Messages, &websock.ChatMessage{
	//		Sender:    message.Sender,
	//		Timestamp: message.Timestamp,
	//		Message:   message.MessageContent[0].Content})
	//}

	// TODO send chatinfo message to user registrating

	// Add user to subscribers
	room.subscribers[username] = struct {
		sink   chan<- websock.Message
		pubKey *rsa.PublicKey
	}{
		sink,
		pubKey,
	}
}

// trySendMsg tries to send an event to a specific user and removes the user
// if its unable to send the message
func (room *ChatRoom) trySendMsg(username string, sink chan<- websock.Message, evt websock.Message) bool {
	select {
	// Try to send message to client sink
	case sink <- evt:
		return true
	// TODO perhaps change, will currently close connection to all
	// clients which do not have room for another event
	default:
		close(sink)
		delete(room.subscribers, username)
		return false
	}
}
