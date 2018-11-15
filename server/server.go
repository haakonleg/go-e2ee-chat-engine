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
	Db *mdb.Database
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
	}

	return s
}

// Packet is a websocket message with an included error message which might
// come from a closed connection
type Packet struct {
	Msg websock.Message
	Err error
}

// WebsockHandler is the handler for the server websocket when a client initially connects.
// It handles messages from an unauthenticated client.
func (s *Server) WebsockHandler(ws *websocket.Conn) {
	log.Printf("Client connected: %s\n", ws.Request().RemoteAddr)

	pinger, pongCount := s.Pinger(ws)

	// Enter unauthenticated message loop
	user, err := s.NoAuthHandler(ws, pongCount)
	if err == nil {
		// Enter authenticated message loop
		s.AuthedHandler(ws, user, pongCount)
	}

	pinger.Stop()
	ws.Close()
	log.Printf("Client disconnected: %s\n", ws.Request().RemoteAddr)
}

// NoAuthHandler handles websocket messages from an unauthenticated client
// This function returns true if the client was authenticated, or false
// if the client disconnected without authenticating as a user
func (s *Server) NoAuthHandler(ws *websocket.Conn, pongCount *int64) (*User, error) {
	// Listen for messages from unauthenticated clients
	for {
		msg := new(websock.Message)
		if err := websock.Msg.Receive(ws, msg); err != nil {
			log.Println(err)
			return nil, err
		}

		// Check message type and forward to appropriate handlers
		switch msg.Type {
		case websock.RegisterUser:
			if ValidateRegisterUser(ws, msg.Message.(*websock.RegisterUserMessage)) {
				s.RegisterUser(ws, msg.Message.(*websock.RegisterUserMessage))
			}
		case websock.LoginUser:
			user, err := s.LoginUser(ws, msg.Message.(string))
			if err == nil {
				return user, nil
			}
		case websock.Pong:
			log.Printf("Receive pong from %s", ws.Request().RemoteAddr)
			atomic.AddInt64(pongCount, 1)
		}
	}
}

// AuthedHandler handles websocket messages from authenticated clients
func (s *Server) AuthedHandler(ws *websocket.Conn, user *User, pongCount *int64) {
	// Listen for messages from authenticated clients
	for {
		msg := new(websock.Message)
		if err := websock.Msg.Receive(ws, msg); err != nil {
			log.Println(err)
			break
		}

		// Check message type and forward to appropriate handlers
		switch msg.Type {
		case websock.CreateChatRoom:
			s.CreateChatRoom(ws, user, msg.Message.(*websock.CreateChatRoomMessage))
		case websock.GetChatRooms:
			s.GetChatRooms(ws)
		case websock.JoinChat:
			s.JoinChat(ws, user, msg.Message.(*websock.JoinChatMessage))
		case websock.SendChat:
			s.ReceiveChatMessage(ws, user, msg.Message.(*websock.SendChatMessage))
		case websock.LeaveChat:
			s.ClientLeftChat(ws, user)
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
