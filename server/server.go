package server

import (
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// Config describes the server configuration, where the listening port,
// name of the mongoDB database used by the server, and the mongoDB address
type Config struct {
	DBName   string
	MongoURL string
}

// Server contains the context of the chat engine server
type Server struct {
	Config
	Db    *mdb.Database
	Users Users
}

// CreateServer creates a new instance of the server using the config
func CreateServer(config Config) *Server {
	// Connect to the database
	db, err := mdb.CreateConnection(config.MongoURL, config.DBName)
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		Config: config,
		Db:     db,
		Users:  Users{data: make(map[*websocket.Conn]*User, 0)},
	}

	return s
}

// AddClient adds a new client to Users
func (s *Server) AddClient(ws *websocket.Conn, user *User) {
	if !s.Users.Insert(ws, user) {
		log.Print("Websocket connection is already associated with a user")
	}
}

// RemoveClient removes a client from the ConnectedClients map
func (s *Server) RemoveClient(ws *websocket.Conn) {
	user, ok := s.Users.Get(ws)
	if !ok || user == nil {
		log.Print("Websocket was not associated with a user")
		return
	}
	user.Lock()
	defer user.Unlock()

	if user.ChatRoom != "" {
		go s.ClientLeftChat(ws)
	}
	s.Users.Remove(ws)
}

// WebsockHandler is the handler for the server websocket
// it handles messages from a single client
func (s *Server) WebsockHandler(ws *websocket.Conn) {
	s.AddClient(ws, nil)
	log.Printf("Client connected: %s. Total connected: %d", ws.Request().RemoteAddr, s.Users.Len())

	// Listen for messages
	for {
		msg := new(websock.Message)
		if err := websocket.JSON.Receive(ws, msg); err != nil {
			break
		}

		// Check message type and forward to appropriate handlers
		switch msg.Type {
		case websock.RegisterUser:
			s.RegisterUser(ws, msg)
		case websock.LoginUser:
			s.LoginUser(ws, msg)
		case websock.CreateChatRoom:
			s.CreateChatRoom(ws, msg)
		case websock.GetChatRooms:
			s.GetChatRooms(ws)
		case websock.JoinChat:
			s.JoinChat(ws, msg)
		case websock.SendChat:
			s.ReceiveChatMessage(ws, msg)
		case websock.LeaveChat:
			s.ClientLeftChat(ws)
		default:
			websock.InvalidMessage(ws)
		}
	}

	s.RemoveClient(ws)
	log.Printf("Client disconnected: %s. Total connected: %d\n", ws.Request().RemoteAddr, s.Users.Len())
}
