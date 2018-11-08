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

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

const (
	authKeyLen = 64
)

type User struct {
	Username  string
	AuthKey   []byte
	EncKey    []byte
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
	if err := s.Db.Insert(mdb.Users, []interface{}{user}); err != nil {
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
	newUser, err := NewUser(s.Db, username)
	if err != nil {
		websock.SendMessage(ws, websock.Error, "User does not exist", websock.String)
		return
	}

	// Send auth challenge
	websock.SendMessage(ws, websock.AuthChallenge, newUser.EncKey, websock.Bytes)

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
func NewUser(db *mdb.Database, username string) (*User, error) {
	// Retrieve user from DB
	query := bson.M{"username": username}

	user := new(mdb.User)
	if err := db.FindOne(mdb.Users, query, nil, user); err != nil {
		log.Println(err)
		return nil, err
	}

	// Unmarshal public key
	pubKey, err := util.UnmarshalPublic(user.PublicKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Generate auth challenge
	encKey, authKey := GenAuthChallenge(pubKey)

	return &User{
		Username:  username,
		AuthKey:   authKey,
		EncKey:    encKey,
		PublicKey: pubKey}, nil
}

// CheckAuth cheks that the recieved authentication token matches the expected token for the user
func (s *Server) CheckAuth(ws *websocket.Conn, authKey []byte) bool {
	user, ok := s.ConnectedClients[ws]
	if !ok || user == nil {
		websock.SendMessage(ws, websock.Error, "Not logged in", websock.String)
	}

	if user.KeyMatches(authKey) {
		return true
	}

	websock.SendMessage(ws, websock.Error, "Invalid auth key", websock.String)
	return false
}

// GenAuthChallenge generates a random authentication key, and encrypts it with the given public key
// returns the encrypted and the original auth key
func GenAuthChallenge(pubKey *rsa.PublicKey) ([]byte, []byte) {
	authKey := make([]byte, authKeyLen)
	rand.Read(authKey)
	encKey, _ := rsa.EncryptPKCS1v15(rand.Reader, pubKey, authKey)
	return encKey, authKey
}
