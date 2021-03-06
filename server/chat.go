package server

import (
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// CreateChatRoom creates a new chat room, and adds it to the database
func (s *Server) CreateChatRoom(ws *websocket.Conn, msg *websock.CreateChatRoomMessage) {
	user, ok := s.Users.Get(ws)
	if !ok || user == nil {
		log.Print("Websocket was not associated with a user")
		return
	}
	user.Lock()
	defer user.Unlock()

	// Add the chat room to the database
	chat := mdb.NewChat(msg.Name, msg.Password, msg.IsHidden)
	if err := s.Db.Insert(mdb.ChatRooms, chat); err != nil {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "Error creating chat room"})
		return
	}

	websock.Send(ws, &websock.Message{Type: websock.OK, Message: "Chat room created"})
}

// GetChatRooms returns all non-hidden chat rooms to the websocket client
func (s *Server) GetChatRooms(ws *websocket.Conn) {
	// Get chat rooms from the database (which are not hidden), and add it to the struct
	results := make([]*mdb.Chat, 0)
	if err := s.Db.FindAll(mdb.ChatRooms, bson.M{"is_hidden": false}, nil, &results); err != nil {
		log.Println(err)
		return
	}

	response := &websock.GetChatRoomsResponseMessage{
		TotalConnected: s.Users.Len(),
		Rooms:          make([]websock.Room, 0, len(results))}

	for _, room := range results {
		response.Rooms = append(response.Rooms, websock.Room{
			Name:        room.Name,
			HasPassword: len(room.PasswordHash) != 0,
			OnlineUsers: s.Users.LenInChat(room.Name)})
	}

	websock.Send(ws, &websock.Message{Type: websock.GetChatRoomsResponse, Message: response})
}

// JoinChat assigns a client to a chat room
func (s *Server) JoinChat(ws *websocket.Conn, msg *websock.JoinChatMessage) {
	// Check that user is logged in
	user, ok := s.Users.Get(ws)
	if !ok || user == nil {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "Not logged in"})
		return
	}
	user.Lock()
	defer user.Unlock()

	// Check that user is not already in a chat room
	if user.ChatRoom != "" {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "You are already in a chat room"})
		return
	}

	// Retrieve the chat room from database
	chat := new(mdb.Chat)
	if err := s.Db.FindOne(mdb.ChatRooms, bson.M{"name": msg.Name}, nil, chat); err != nil {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "This chat room does not exist"})
		return
	}

	// Verify password (if necessary)
	if len(chat.PasswordHash) != 0 && !chat.ValidPassword(msg.Password) {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "Invalid password"})
		return
	}

	// Add user to chat room
	user.ChatRoom = msg.Name
	websock.Send(ws, &websock.Message{Type: websock.OK, Message: "Joined chat"})

	s.ClientJoinedChat(ws, user, msg.Name)
}

// ClientJoinedChat is called when a client joins a chat room, it adds the username of the client
// to the map of chat rooms and the chat room name to the User object, to be able to keep track of this
// Then info about the chat room, and messages for this user is sent to the client
func (s *Server) ClientJoinedChat(ws *websocket.Conn, user *User, chatName string) {
	// Create response object, send the client list of users, and messages sent that this user can decrypt
	chatInfo := &websock.ChatInfoMessage{
		MyUsername: user.Username,
		Users: []websock.User{{
			Username:  user.Username,
			PublicKey: util.MarshalPublic(user.PublicKey)}},
		Messages: make([]*websock.ChatMessage, 0)}

	s.Users.ForEachInChat(chatName, func(client *websocket.Conn, otherUser *User) {
		if otherUser == user {
			return
		}
		otherUser.Lock()
		defer otherUser.Unlock()
		chatInfo.Users = append(chatInfo.Users, websock.User{
			Username:  otherUser.Username,
			PublicKey: util.MarshalPublic(otherUser.PublicKey)})
	})

	// Add the chat messages addressed to this user
	for _, message := range s.FindMessagesForUser(user.Username, chatName) {
		// Check if the message actually has the encrypted message
		if len(message.MessageContent) == 0 {
			continue
		}

		chatInfo.Messages = append(chatInfo.Messages, &websock.ChatMessage{
			Sender:    message.Sender,
			Timestamp: message.Timestamp,
			Message:   message.MessageContent[0].Content})
	}

	go websock.Send(ws, &websock.Message{Type: websock.ChatInfo, Message: chatInfo})

	// Notify other clients in the chat that a new user has joined
	s.NotifyUserJoined(user, chatName)
}

// ClientLeftChat is called when a client leaves a chat room, it removes the username of the client
// from the map of chat rooms and the chat room name from the User object. Other clients in the chat
// will be notfied that this user left the chat as well
func (s *Server) ClientLeftChat(ws *websocket.Conn) {
	user, ok := s.Users.Get(ws)
	if !ok || user == nil {
		log.Print("Websocket was not associated with a user")
		return
	}
	user.Lock()
	defer user.Unlock()

	chatName := user.ChatRoom
	username := user.Username
	user.ChatRoom = ""

	go websock.Send(ws, &websock.Message{Type: websock.UserLeft, Message: username})
	// Notify clients that this user left the chat
	go s.NotifyUserLeft(username, chatName)
}

// FindMessagesForUser finds all chat messages with a specific user as recipient in a specific chat room
func (s *Server) FindMessagesForUser(username, chatName string) []*mdb.Message {
	query := bson.M{
		"chat_name": chatName}

	selector := bson.M{
		"timestamp": 1,
		"sender":    1,
		"message_content": bson.M{
			"$elemMatch": bson.M{"recipient": username}},
	}

	result := make([]*mdb.Message, 0)
	if err := s.Db.FindAll(mdb.Messages, query, selector, &result); err != nil {
		return nil
	}

	return result
}

// ReceiveChatMessage is called when the server receives a chat message from a client that is in a chat room
func (s *Server) ReceiveChatMessage(ws *websocket.Conn, msg *websock.SendChatMessage) {
	user, ok := s.Users.Get(ws)
	if !ok || user == nil {
		log.Print("Websocket was not associated with a user")
		return
	}
	user.Lock()
	defer user.Unlock()

	// Check that the client is actually in a chat room
	if user.ChatRoom == "" {
		websock.Send(ws, &websock.Message{Type: websock.Error, Message: "You are not in a chat room"})
		return
	}

	websock.Send(ws, &websock.Message{Type: websock.OK, Message: "Message sent"})

	// Notify everyone in the chat room about the new chat message, and store the message in the database
	timestamp := util.NowMillis()
	go s.NotifyChatMessage(user.Username, user.ChatRoom, timestamp, msg.EncryptedContent)
	go s.AddMessageToDB(user.Username, user.ChatRoom, timestamp, msg.EncryptedContent)
}

// NotifyChatMessage notifies all clients in a chat room about a new chat message
func (s *Server) NotifyChatMessage(sender string, chatName string, timestamp int64, encryptedContent map[string][]byte) {

	// Notify the clients in the chat room
	s.Users.ForEachInChat(chatName, func(client *websocket.Conn, recipent *User) {
		recipent.Lock()
		defer recipent.Unlock()
		msg := &websock.ChatMessage{
			Sender:    sender,
			Timestamp: timestamp,
			Message:   encryptedContent[recipent.Username]}

		go websock.Send(client, &websock.Message{Type: websock.ChatMessageReceived, Message: msg})
	})
}

// NotifyUserJoined notifies all clients in a chat room that a new user has joined the chat room
func (s *Server) NotifyUserJoined(user *User, chatName string) {
	msg := &websock.User{
		Username:  user.Username,
		PublicKey: util.MarshalPublic(user.PublicKey)}

	go s.Users.ForEachInChat(chatName, func(client *websocket.Conn, otherUser *User) {
		if otherUser == user {
			return
		}
		otherUser.Lock()
		defer otherUser.Unlock()
		go websock.Send(client, &websock.Message{Type: websock.UserJoined, Message: msg})
	})
}

// NotifyUserLeft notifies all clients in a chat room that a user left the chat room
func (s *Server) NotifyUserLeft(username, chatName string) {
	// Get all clients in the chat room
	s.Users.ForEachInChat(chatName, func(client *websocket.Conn, _ *User) {
		go websock.Send(client, &websock.Message{Type: websock.UserLeft, Message: username})
	})
}

// AddMessageToDB inserts a chat message into the database
func (s *Server) AddMessageToDB(username, chatName string, timestamp int64, encryptedContent map[string][]byte) {
	chatMessage := mdb.NewMessage(chatName, timestamp, username)

	for recipient, encryptedMessage := range encryptedContent {
		msg := mdb.MessageContent{
			Recipient: recipient,
			Content:   encryptedMessage}
		chatMessage.MessageContent = append(chatMessage.MessageContent, msg)
	}

	if err := s.Db.Insert(mdb.Messages, chatMessage); err != nil {
		log.Println(err)
		return
	}
}
