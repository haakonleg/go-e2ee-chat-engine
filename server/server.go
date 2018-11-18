package server

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// Config describes the server configuration, where the listening port,
// name of the mongoDB database used by the server, and the mongoDB address
type Config struct {
	DBName    string
	MongoURL  string
	Keepalive int
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
	user, ok := s.Users.Remove(ws)
	if !ok {
		log.Print("Websocket was not in users-map")
		return
	}
	if user == nil {
		log.Print("Websocket was not associated with a user")
	} else {
		user.Lock()
		defer user.Unlock()
		if user.ChatRoom != "" {
			s.ClientLeftChat(ws)
		} else {
			log.Print("User was not associated with a chatroom")
		}
	}
}

// WebsockHandler is the handler for the server websocket when a client initially connects.
// It handles messages from an unauthenticated client.
func (s *Server) WebsockHandler(ws *websocket.Conn) {
	s.AddClient(ws, nil)
	log.Printf("Client connected: %s. Total connected: %d", ws.Request().RemoteAddr, s.Users.Len())

	pinger, pongCount := s.Pinger(ws)

	// Enter unauthenticated message loop
	if s.NoAuthHandler(ws, pongCount) {
		// Enter authenticated message loop
		s.AuthedHandler(ws, pongCount)
	}

	pinger.Stop()
	ws.Close()
	s.RemoveClient(ws)
	log.Printf("Client disconnected: %s. Total connected: %d\n", ws.Request().RemoteAddr, s.Users.Len())
}

// NoAuthHandler handles websocket messages from an unauthenticated client
// This function returns true if the client was authenticated, or false
// if the client disconnected without authenticating as a user
func (s *Server) NoAuthHandler(ws *websocket.Conn, pongCount *int64) bool {
	// Listen for messages from unauthenticated clients
	for {
		msg := new(websock.Message)
		if err := websock.Msg.Receive(ws, msg); err != nil {
			log.Println(err)
			return false
		}

		// Check message type and forward to appropriate handlers
		switch msg.Type {
		case websock.RegisterUser:
			if ValidateRegisterUser(ws, msg.Message.(*websock.RegisterUserMessage)) {
				s.RegisterUser(ws, msg.Message.(*websock.RegisterUserMessage))
			}
		case websock.LoginUser:
			if s.LoginUser(ws, msg.Message.(string)) {
				return true
			}
		case websock.Pong:
			log.Printf("Receive pong from %s", ws.Request().RemoteAddr)
			atomic.AddInt64(pongCount, 1)
		}
	}
}

// AuthedHandler handles websocket messages from authenticated clients
func (s *Server) AuthedHandler(ws *websocket.Conn, pongCount *int64) {
	// Listen for messages from authenticated clients
	for {
		msg := new(websock.Message)
		if err := websock.Msg.Receive(ws, msg); err != nil {
			log.Println(err)
			break
		}

		switch msg.Type {
		case websock.CreateChatRoom:
			if ValidateCreateChatRoom(ws, msg.Message.(*websock.CreateChatRoomMessage)) {
				s.CreateChatRoom(ws, msg.Message.(*websock.CreateChatRoomMessage))
			}
		case websock.GetChatRooms:
			s.GetChatRooms(ws)
		case websock.JoinChat:
			s.JoinChat(ws, msg.Message.(*websock.JoinChatMessage))
		case websock.SendChat:
			s.ReceiveChatMessage(ws, msg.Message.(*websock.SendChatMessage))
		case websock.LeaveChat:
			s.ClientLeftChat(ws)
		case websock.Pong:
			log.Printf("Receive pong from %s", ws.Request().RemoteAddr)
			atomic.AddInt64(pongCount, 1)
		}
	}
}

// Pinger sends a ping message to the client in the interval specified in Keepalive in the ServerConfig
// If no pongs were received during the elapsed time, the server will close the client connection.
func (s *Server) Pinger(ws *websocket.Conn) (*time.Ticker, *int64) {
	ticker := time.NewTicker(time.Duration(s.Keepalive) * time.Second)
	pongCount := int64(1)

	go func() {
		for range ticker.C {
			if atomic.LoadInt64(&pongCount) == 0 {
				log.Printf("Client %s did not respond to ping in time", ws.Request().RemoteAddr)
				ws.Close()
				return
			}

			websock.Msg.Send(ws, &websock.Message{Type: websock.Ping})
			atomic.StoreInt64(&pongCount, 0)
		}
	}()

	return ticker, &pongCount
}
