package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

const (
	authKeyLen = 64
)

// Users is a threadsafe connection between a websocket connection and a user
//
// The mutex must be held when accessing or modifying the map
type Users struct {
	sync.Mutex
	// The currently connected clients, if a connected client has logged in
	// the key (websocket.Conn pointer) will refer to a user.User object, else nil
	data map[*websocket.Conn]*User
}

// Get gets the User of a connected websocket client
//
// Returns true on success and false on missing user
func (users *Users) Get(ws *websocket.Conn) (user *User, ok bool) {
	users.Lock()
	defer users.Unlock()
	user, ok = users.data[ws]
	return
}

// Remove deletes the connection between a websocket and a user
func (users *Users) Remove(ws *websocket.Conn) (user *User, ok bool) {
	users.Lock()
	defer users.Unlock()

	user, ok = users.data[ws]
	if ok {
		delete(users.data, ws)
	}
	return
}

// Insert adds the given User to the collection indexed by the websocket
// connection
//
// Returns true on success and false on already existing association between
// socket and user
func (users *Users) Insert(ws *websocket.Conn, user *User) bool {
	users.Lock()
	defer users.Unlock()

	// This connection already has an associated user
	if user, ok := users.data[ws]; ok && user != nil {
		return false
	}
	users.data[ws] = user
	return true
}

// ForEach performs the given function on all stored users
func (users *Users) ForEach(f func(*websocket.Conn, *User)) {
	users.Lock()
	defer users.Unlock()
	for ws, user := range users.data {
		f(ws, user)
	}
}

// ForEachInChat performs the given function for every user which is in the
// given chat
func (users *Users) ForEachInChat(chatName string, f func(*websocket.Conn, *User)) {
	users.Lock()
	defer users.Unlock()
	for ws, user := range users.data {
		if user == nil {
			continue
		}
		if user.ChatRoom == chatName {
			f(ws, user)
		}
	}
}

// Len gets the amount of registered users
func (users *Users) Len() int {
	users.Lock()
	defer users.Unlock()
	return len(users.data)
}

// LenInChat gets the amount of registered users in a given chat
func (users *Users) LenInChat(chatName string) (amount int) {
	users.Lock()
	defer users.Unlock()
	for _, user := range users.data {
		if user == nil {
			continue
		}
		user.Lock()
		if user.ChatRoom == chatName {
			amount++
		}
		user.Unlock()
	}
	return
}

// User contains user data and a mutex to enable threadsafe access without
// copying
//
// The mutex must be held when accessing or modifying fields
type User struct {
	sync.Mutex
	Username  string
	AuthKey   []byte
	PublicKey *rsa.PublicKey
	ChatRoom  string
}

// ValidateUsername checks that a username fulfills certain requirements. It
// returns an error with a descriptive message if does not uphold requirements.
//
// FIXME should probably be moved to a different file/module when validation of
// other messages is improved
func ValidateUsername(username string) error {
	if username == "" {
		return errors.New("Username cannot be empty")
	}
	if len(username) < 3 {
		return errors.New("Username has to contain at least 3 characters")
	}
	if len(username) > 20 {
		return errors.New("Username cannot contain more than 20 characters")
	}
	return nil
}

// KeyMatches checks that an authentication key matches the one for this user
func (u *User) KeyMatches(authKey []byte) bool {
	return bytes.Compare(u.AuthKey, authKey) == 0
}

// RegisterUser registers a new user, and adds it to the database
func (s *Server) RegisterUser(ws *websocket.Conn, msg *websock.Message) {
	regUserMsg := new(websock.RegisterUserMessage)
	if err := json.Unmarshal(msg.Message, regUserMsg); err != nil {
		websock.InvalidFormat(ws)
		return
	}
	regUserMsg.Username = strings.TrimSpace(regUserMsg.Username)

	if err := ValidateUsername(regUserMsg.Username); err != nil {
		websock.SendMessage(ws, websock.Error, "Invalid username, "+err.Error(), websock.String)
		return
	}

	// Add new user to database
	user := mdb.NewUser(regUserMsg.Username, regUserMsg.PublicKey)
	if err := s.Db.Insert(mdb.Users, user); err != nil {
		websock.SendMessage(ws, websock.Error, "Error registering user", websock.String)
		return
	}

	websock.SendMessage(ws, websock.MessageOK, "User registered", websock.String)
}

// LoginUser authenticates a user using a randomly generated authentication token
// This token is encrypted with the public key of the username the client is trying to log in as
// The client is then expected to respond with the correct decrypted token
func (s *Server) LoginUser(ws *websocket.Conn, msg *websock.Message) {
	username := string(msg.Message)

	// Create new user object
	newUser, encKey, err := NewUser(s.Db, username)
	if err != nil {
		websock.SendMessage(ws, websock.Error, "User does not exist", websock.String)
		return
	}

	// Send auth challenge
	websock.SendMessage(ws, websock.AuthChallenge, encKey, websock.Bytes)

	// Recieve auth challenge response
	res, err := websock.GetResponse(ws)
	if err != nil {
		log.Println(err)
		return
	}

	// Check that the recieved decrypted key matches the original auth key
	if newUser.KeyMatches(res.Message) {
		fmt.Printf("Client %s authenticated as user %s\n", ws.Request().RemoteAddr, newUser.Username)
		s.AddClient(ws, newUser)
		websock.SendMessage(ws, websock.MessageOK, "Logged in", websock.String)
	} else {
		websock.SendMessage(ws, websock.Error, "Invalid auth key", websock.String)
	}
}

// NewUser creates a new user object for a connected client, with the username, generated (temporary) authentication
// key and the encrypted version of the key. A random byte slice is generated and encrypted with the users public key, the user
// is expected to send in response the decrypted string
func NewUser(db *mdb.Database, username string) (*User, []byte, error) {
	// Retrieve user from DB
	query := bson.M{"username": username}

	user := new(mdb.User)
	if err := db.FindOne(mdb.Users, query, nil, user); err != nil {
		log.Println(err)
		return nil, nil, err
	}

	// Unmarshal public key
	pubKey, err := util.UnmarshalPublic(user.PublicKey)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	// Generate auth challenge
	encKey, authKey := GenAuthChallenge(pubKey)

	return &User{
		Username:  username,
		AuthKey:   authKey,
		PublicKey: pubKey}, encKey, nil
}

// GenAuthChallenge generates a random authentication key, and encrypts it with the given public key
// returns the encrypted and the original auth key
func GenAuthChallenge(pubKey *rsa.PublicKey) ([]byte, []byte) {
	authKey := make([]byte, authKeyLen)
	rand.Read(authKey)
	encKey, _ := rsa.EncryptPKCS1v15(rand.Reader, pubKey, authKey)
	return encKey, authKey
}
