package server

import (
	"encoding/json"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// CreateChatRoom creates a new chat room, and adds it to the database
func (s *Server) CreateChatRoom(ws *websocket.Conn, msg *websock.Message) {
	createChatRoomMsg := new(websock.CreateChatRoomMessage)
	if err := json.Unmarshal(msg.Message, createChatRoomMsg); err != nil {
		websock.InvalidFormat(ws)
		return
	}

	// Check that authentication key is correct
	if !s.CheckAuth(ws, createChatRoomMsg.AuthKey) {
		websock.InvalidAuth(ws)
		return
	}

	if len(createChatRoomMsg.Name) < 3 {
		websock.SendMessage(ws, websock.Error, "Name must be 3 characters or longer", websock.String)
		return
	}

	// Add the chat room to the database
	chat := mdb.NewChat(createChatRoomMsg.Name)
	if err := s.Db.Insert(mdb.ChatRooms, []interface{}{chat}); err != nil {
		websock.SendMessage(ws, websock.Error, "Error creating chat room", websock.String)
		return
	}

	websock.SendMessage(ws, websock.MessageOK, "Chat room created", websock.String)
}

// GetChatRooms returns all chat rooms to the websocket client
func (s *Server) GetChatRooms(ws *websocket.Conn) {
	// Get chat rooms from the database and add it to the slice
	results := make([]*mdb.Chat, 0)
	if err := s.Db.FindAll(mdb.ChatRooms, nil, nil, &results); err != nil {
		log.Println(err)
	}

	response := &websock.GetChatRoomsResponseMessage{
		Rooms: make([]websock.Room, 0, len(results))}

	for _, room := range results {
		response.Rooms = append(response.Rooms, websock.Room{
			Name:        room.Name,
			OnlineUsers: len(s.FindClientsInChat(room.Name))})
	}

	websock.SendMessage(ws, websock.GetChatRoomsResponse, response, websock.JSON)
}

// JoinChat assigns a client to a chat room
func (s *Server) JoinChat(ws *websocket.Conn, msg *websock.Message) {
	joinChatMsg := new(websock.JoinChatMessage)
	if err := json.Unmarshal(msg.Message, joinChatMsg); err != nil {
		websock.InvalidFormat(ws)
		return
	}

	// Check that authentication key is correct
	if !s.CheckAuth(ws, joinChatMsg.AuthKey) {
		websock.InvalidAuth(ws)
		return
	}

	// Check that user is not already in a chat room
	s.ccMtx.Lock()
	if s.ConnectedClients[ws].ChatRoom != "" {
		websock.SendMessage(ws, websock.Error, "You are already in a chat room", websock.String)
		return
	}
	s.ccMtx.Unlock()

	// Check that chat room exists
	if !s.Db.DocumentExists(mdb.ChatRooms, bson.M{"name": joinChatMsg.Name}) {
		websock.SendMessage(ws, websock.Error, "This chat room does not exist", websock.String)
		return
	}

	// Add user to chat room
	websock.SendMessage(ws, websock.MessageOK, "Joined chat", websock.String)

	s.ClientJoinedChat(ws, joinChatMsg.Name)
}

// ClientJoinedChat is called when a client joins a chat room, it adds the username of the client
// to the map of chat rooms and the chat room name to the User object, to be able to keep track of this
// Then info about the chat room, and messages for this user is sent to the client
func (s *Server) ClientJoinedChat(ws *websocket.Conn, chatName string) {
	s.ConnectedClients[ws].ChatRoom = chatName

	// Create response object, send the client list of users, and messages sent that this user can decrypt
	chatInfo := &websock.ChatInfoMessage{
		Users:    make([]websock.User, 0),
		Messages: make([]*websock.ChatMessage, 0)}

	for _, client := range s.FindClientsInChat(chatName) {
		u, ok := s.ConnectedClients[client]
		if !ok {
			continue
		}
		chatInfo.Users = append(chatInfo.Users, websock.User{
			Username:  u.Username,
			PublicKey: util.MarshalPublic(u.PublicKey)})
	}

	// Add the chat messages addressed to this user
	for _, message := range s.FindMessagesForUser(s.ConnectedClients[ws].Username, chatName) {
		// Check if the message actually has the encrypted message
		if len(message.MessageContent) == 0 {
			continue
		}

		chatInfo.Messages = append(chatInfo.Messages, &websock.ChatMessage{
			Sender:    message.Sender,
			Timestamp: message.Timestamp,
			Message:   message.MessageContent[0].Content})
	}

	websock.SendMessage(ws, websock.ChatInfo, chatInfo, websock.JSON)

	// Notify other clients in the chat that a new user has joined
	go s.NotifyUserJoined(ws, chatName)
}

// ClientLeftChat is called when a client leaves a chat room, it removes the username of the client
// from the map of chat rooms and the chat room name from the User object. Other clients in the chat
// will be notfied that this user left the chat as well
func (s *Server) ClientLeftChat(ws *websocket.Conn, chatName string) {
	s.ccMtx.Lock()
	defer s.ccMtx.Unlock()

	username := s.ConnectedClients[ws].Username
	s.ConnectedClients[ws].ChatRoom = ""
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

// FindClientsInChat returns every client that is connected to a particular chat room
func (s *Server) FindClientsInChat(chatName string) []*websocket.Conn {
	s.ccMtx.Lock()
	defer s.ccMtx.Unlock()

	clients := make([]*websocket.Conn, 0)
	for client, user := range s.ConnectedClients {
		if user.ChatRoom == chatName {
			clients = append(clients, client)
		}
	}
	return clients
}

// ReceiveChatMessage is called when the server recieves a chat message from a client that is in a chat room
func (s *Server) ReceiveChatMessage(ws *websocket.Conn, msg *websock.Message) {
	sendChatMsg := new(websock.SendChatMessage)
	if err := json.Unmarshal(msg.Message, sendChatMsg); err != nil {
		websock.InvalidFormat(ws)
		return
	}

	// Check that authentication key is correct
	if !s.CheckAuth(ws, sendChatMsg.AuthKey) {
		websock.InvalidAuth(ws)
		return
	}

	// Check that the client is actually in a chat room
	s.ccMtx.Lock()
	username := s.ConnectedClients[ws].Username
	chatName := s.ConnectedClients[ws].ChatRoom
	s.ccMtx.Unlock()

	if chatName == "" {
		websock.SendMessage(ws, websock.Error, "You are not in a chat room", websock.String)
		return
	}

	websock.SendMessage(ws, websock.MessageOK, "Message sent", websock.String)

	// Notify everyone in the chat room about the new chat message, and store the message in the database
	timestamp := util.NowMillis()
	go s.NotifyChatMessage(username, chatName, timestamp, sendChatMsg.EncryptedContent)
	go s.AddMessageToDB(username, chatName, timestamp, sendChatMsg.EncryptedContent)
}

// NotifyChatMessage notifies all clients in a chat room about a new chat message
func (s *Server) NotifyChatMessage(sender string, chatName string, timestamp int64, encryptedContent map[string][]byte) {
	// Get all clients in the chat room
	clients := s.FindClientsInChat(chatName)

	// Notify the clients in the chat room
	for _, client := range clients {
		s.ccMtx.Lock()
		recipient := s.ConnectedClients[client].Username
		s.ccMtx.Unlock()

		msg := &websock.ChatMessage{
			Sender:    sender,
			Timestamp: timestamp,
			Message:   encryptedContent[recipient]}

		websock.SendMessage(client, websock.ChatMessageReceived, msg, websock.JSON)
	}
}

// NotifyUserJoined notifies all clients in a chat room that a new user has joined the chat room
func (s *Server) NotifyUserJoined(joined *websocket.Conn, chatName string) {
	// Get user object
	s.ccMtx.Lock()
	username := s.ConnectedClients[joined].Username
	pubKey := s.ConnectedClients[joined].PublicKey
	s.ccMtx.Unlock()

	msg := &websock.User{
		Username:  username,
		PublicKey: util.MarshalPublic(pubKey)}

	for _, client := range s.FindClientsInChat(chatName) {
		if client != joined {
			websock.SendMessage(client, websock.UserJoined, msg, websock.JSON)
		}
	}
}

// NotifyUserLeft notifies all clients in a chat room that a user left the chat room
func (s *Server) NotifyUserLeft(username, chatName string) {
	// Get all clients in the chat room
	for _, client := range s.FindClientsInChat(chatName) {
		websock.SendMessage(client, websock.UserLeft, username, websock.String)
	}
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

	if err := s.Db.Insert(mdb.Messages, []interface{}{chatMessage}); err != nil {
		log.Println(err)
	}
}
