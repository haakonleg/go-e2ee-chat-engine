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
	Db          *mdb.Database
	ChatManager *ChatRoomManager
}

// CreateServer creates a new instance of the server using the config
func CreateServer(config Config) *Server {
	// Connect to the database
	db, err := mdb.CreateConnection(config.MongoURL, config.DBName)
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		Config:      config,
		Db:          db,
		ChatManager: NewChatRoomManager(),
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

	// Turn received packets into a channel to enable the use of select
	packets := make(chan Packet)
	getPacket := make(chan struct{})

	// Start receiving JSON packets and send them to the packets channel
	go func(ws *websocket.Conn, pks chan<- Packet, wait <-chan struct{}) {
		var pk Packet

		for {
			// Only continue when were asked too. A lot of handlers
			// send/receive on the same websocket, so we do not want to
			// intervene
			<-wait
			pk.Err = websocket.JSON.Receive(ws, &pk.Msg)
			pks <- pk

			// TODO does an error always mean that the connection is closed? If
			// not this should be done differently
			if pk.Err != nil {
				break
			}
		}
	}(ws, packets, getPacket)

	// Enter unauthenticated message loop
	user, err := s.NoAuthHandler(ws, packets, getPacket, pongCount)
	if err == nil {
		// Enter authenticated message loop
		s.AuthedHandler(ws, user, packets, getPacket, pongCount)
	}

	pinger.Stop()
	ws.Close()
	log.Printf("Client disconnected: %s\n", ws.Request().RemoteAddr)
}

// NoAuthHandler handles websocket messages from an unauthenticated client
// This function returns true if the client was authenticated, or false
// if the client disconnected without authenticating as a user
func (s *Server) NoAuthHandler(ws *websocket.Conn, packets <-chan Packet, getPacket chan<- struct{}, pongCount *int64) (*User, error) {
	var packet Packet
	// Listen for messages from unauthenticated clients
	for {
		// Tell packet receiver that we want another packet
		getPacket <- struct{}{}
		packet = <-packets
		if packet.Err != nil {
			log.Printf("Client disconnected before authenticating: %s\n", packet.Err)
			return nil, packet.Err
		}
		// Check message type and forward to appropriate handlers
		switch packet.Msg.Type {
		case websock.RegisterUser:
			if ValidateRegisterUser(ws, packet.Msg.Message.(*websock.RegisterUserMessage)) {
				s.RegisterUser(ws, packet.Msg.Message.(*websock.RegisterUserMessage))
			}
		case websock.LoginUser:
			if user, err := s.LoginUser(ws, packet.Msg.Message.(string)); err == nil {
				return user, nil
			}
		case websock.Pong:
			log.Printf("Receive pong from %s", ws.Request().RemoteAddr)
			atomic.AddInt64(pongCount, 1)
		}
	}
}

// AuthedHandler handles websocket messages from authenticated clients
func (s *Server) AuthedHandler(ws *websocket.Conn, user *User, packets <-chan Packet, getPacket chan<- struct{}, pongCount *int64) {
	if user == nil {
		log.Println("Connection moved to authenticated-loop without having a user object")
		return
	}

	var packet Packet
	// Listen for messages before user is authenticated
	for {
		// Tell packet receiver that we want another packet
		getPacket <- struct{}{}
		packet = <-packets
		if packet.Err != nil {
			log.Printf("User (%s) disconnected : %s\n", user.Username, packet.Err)
			break
		}

		// Check message type and forward to appropriate handlers
		switch packet.Msg.Type {
		case websock.CreateChatRoom:
			s.CreateChatRoom(ws, user, packet.Msg.Message.(*websock.CreateChatRoomMessage))
		case websock.GetChatRooms:
			s.GetChatRooms(ws)
		case websock.JoinChat:
			s.JoinChat(ws, user, packet.Msg.Message.(*websock.JoinChatMessage))
		case websock.SendChat:
			s.ReceiveChatMessage(ws, user, packet.Msg.Message.(*websock.SendChatMessage))
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
