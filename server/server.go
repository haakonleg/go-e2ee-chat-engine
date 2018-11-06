package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

// Config describes the server configuration, where the listening port,
// name of the mongoDB database used by the server, and the mongoDB address
type Config struct {
	ListenPort string
	DBName     string
	MongoURL   string
}

// Server contains the context of the chat engine server
type Server struct {
	Config
	Db *mdb.Database

	// The currently connected clients, if a connected client has logged in
	// the key (websocket.Conn pointer) will refer to a user.User object, else nil
	ccMtx            sync.Mutex
	ConnectedClients map[*websocket.Conn]*User
}

// CreateServer creates a new instance of the server using the config
func CreateServer(config Config) *Server {
	// Connect to the database
	db, err := mdb.CreateConnection(config.MongoURL, config.DBName)
	if err != nil {
		log.Fatal(err)
	}

	s := &Server{
		Config:           config,
		Db:               db,
		ConnectedClients: make(map[*websocket.Conn]*User, 0)}

	return s
}

// Start starts the HTTP server and listens for incoming websocket connections
func (s *Server) Start() {
	// Listen for websocket connections
	log.Println("Listening for incoming connections...")
	http.ListenAndServe(":"+s.ListenPort, websocket.Handler(s.WebsockHandler))
}

// AddClient adds a new client to the ConnectedClients map, go maps are not thread-safe
// so access must be synchronized
func (s *Server) AddClient(ws *websocket.Conn, user *User) {
	s.ccMtx.Lock()
	s.ConnectedClients[ws] = user
	s.ccMtx.Unlock()
}

// RemoveClient removes a client from the ConnectedClients map
func (s *Server) RemoveClient(ws *websocket.Conn) {
	s.ccMtx.Lock()
	delete(s.ConnectedClients, ws)
	s.ccMtx.Unlock()
}

// WebsockHandler is the handler for the server websocket
// it handles messages from a single client
func (s *Server) WebsockHandler(ws *websocket.Conn) {
	s.AddClient(ws, nil)
	log.Printf("Client connected: %s. Total connected: %d", ws.Request().RemoteAddr, len(s.ConnectedClients))

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
		default:
			websock.InvalidMessage(ws)
		}
	}

	s.RemoveClient(ws)
	log.Printf("Client disconnected: %s. Total connected: %d\n", ws.Request().RemoteAddr, len(s.ConnectedClients))
}
