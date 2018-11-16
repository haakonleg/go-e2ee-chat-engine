package server

import (
	"crypto/rsa"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"log"
)

// ChatRoom contains the information about the clients in a chatroom and has
// the responsibility to distribute messages to the corrent correspondent
type ChatRoom struct {
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
	// The broadcast where incoming chat messages are recived
	broadcast chan struct {
		sender     string
		timestamp  int64
		encContent map[string][]byte
	}
}

// Run performs the event handling loop of the chatroom
func (room *ChatRoom) Run() {
	for {
		select {
		// Send chat message to all subscribers
		case msg := <-room.broadcast:
			room.broadcastChatMessage(msg.sender, msg.timestamp, msg.encContent)
		// Register a new subscriber
		case reg := <-room.register:
			room.registerSubscriber(reg.username, reg.pubKey, reg.sink)
		}
	}
}

// Register registers a user to receive events from the chatroom
func (room *ChatRoom) Register(username string, pubKey *rsa.PublicKey) <-chan websock.Message {
	sink := make(chan websock.Message, 3)
	room.register <- struct {
		username string
		sink     chan<- websock.Message
		pubKey   *rsa.PublicKey
	}{username, sink, pubKey}
	return sink
}

// Broadcast sends chat message to all subscribers
func (room *ChatRoom) Broadcast(sender string, timestamp int64, encContent map[string][]byte) {
	room.broadcast <- struct {
		sender     string
		timestamp  int64
		encContent map[string][]byte
	}{
		sender,
		timestamp,
		encContent,
	}
}

// broadcaseChatMessage sends a received chat message to all subscribers
func (room *ChatRoom) broadcastChatMessage(sender string, timestamp int64, encContent map[string][]byte) {
	// Template for a message
	chatmsg := websock.ChatMessage{
		Sender:    sender,
		Timestamp: timestamp,
	}

	// Iterate over all subscribers
	for username, user := range room.subscribers {

		// Get the encrypted content for the current user
		content, ok := encContent[username]
		if !ok {
			// TODO properly handle that message did not include a
			// cipher-message for specific user
			continue
		}

		chatmsg.Message = content
		msg := websock.Message{websock.ChatMessageReceived, chatmsg}
		room.trySendEvent(username, user.sink, msg)
	}
}

// registerSubscriber registers a subscriber to the chatroom
func (room *ChatRoom) registerSubscriber(username string, pubKey *rsa.PublicKey, sink chan<- websock.Message) {

	if _, ok := room.subscribers[username]; ok {
		log.Printf("User (%s) tried to subscribe to a chatroom multiple times\n", username)
		sink <- websock.Message{websock.Error, nil}
		close(sink)
		return
	}

	websockuser := websock.User{
		username,
		util.MarshalPublic(pubKey),
	}

	// Warn all current subscribers that a user joined
	evt := websock.Message{
		websock.UserJoined,
		websockuser,
	}

	// TODO send chatinfo message to user registrating
	chatInfo := &websock.ChatInfoMessage{
		MyUsername: username,
		Users: []websock.User{
			websockuser,
		},
		Messages: make([]*websock.ChatMessage, 0)}

	// Iterate over all subscribers, send message and append to chatusers if
	// successfully sent message
	for otherusername, otheruser := range room.subscribers {
		if room.trySendEvent(otherusername, otheruser.sink, evt) {
			chatInfo.Users = append(chatInfo.Users, websock.User{otherusername, util.MarshalPublic(otheruser.pubKey)})
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

	// Add user to subscribers
	room.subscribers[username] = struct {
		sink   chan<- websock.Message
		pubKey *rsa.PublicKey
	}{
		sink,
		pubKey,
	}
}

// trySendEvent tries to send an event to a specific user and removes the user
// if its unable to send the message
func (room *ChatRoom) trySendEvent(username string, sink chan<- websock.Message, evt websock.Message) bool {
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
