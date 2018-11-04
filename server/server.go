package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/user"
	"golang.org/x/net/websocket"
)

type Config struct {
	ListenPort string
	DBName     string
	MongoURL   string
}

type Server struct {
	Config
	Db *mdb.Database

	// The currently connected clients, if a connected client has logged in
	// the key (websocket.Conn pointer) will refer to a user.User object, else nil
	mapMutex         sync.Mutex
	ConnectedClients map[*websocket.Conn]*user.User
}

func CreateServer(config Config) *Server {
	// Connect to the database
	db, err := mdb.CreateConnection(config.MongoURL, config.DBName)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		Config:           config,
		Db:               db,
		ConnectedClients: make(map[*websocket.Conn]*user.User, 0)}
}

func (s *Server) Start() {
	// Listen for websocket connections
	log.Println("Listening for incoming connections...")
	http.ListenAndServe(":"+s.ListenPort, websocket.Handler(s.WebsockHandler))
}

func (s *Server) AddClient(ws *websocket.Conn, user *user.User) {
	s.mapMutex.Lock()
	s.ConnectedClients[ws] = user
	s.mapMutex.Unlock()
}

func (s *Server) RemoveClient(ws *websocket.Conn) {
	s.mapMutex.Lock()
	delete(s.ConnectedClients, ws)
	s.mapMutex.Unlock()
}
