package server

import (
	"encoding/json"
	"log"

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
	response := new(websock.GetChatRoomsResponse)
	response.Rooms = make([]websock.Room, 0, len(s.ChatRooms))
	for _, room := range s.ChatRooms {
		response.Rooms = append(response.Rooms, websock.Room{
			Name:        room.Name,
			OnlineUsers: len(room.Users)})
	}

	log.Printf("Send chat rooms: %v", response.Rooms)

	websock.SendMessage(ws, websock.ChatRoomsResponse, response, websock.JSON)
}
