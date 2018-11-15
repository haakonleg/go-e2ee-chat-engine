package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"log"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"golang.org/x/net/websocket"
)

const (
	authKeyLen = 64
)

// User contains user data
type User struct {
	Username  string
	AuthKey   []byte
	PublicKey *rsa.PublicKey
	ChatRoom  string
}

// KeyMatches checks that an authentication key matches the one for this user
func (u *User) KeyMatches(authKey []byte) bool {
	return bytes.Equal(u.AuthKey, authKey)
}

// RegisterUser registers a new user, and adds it to the database
func (s *Server) RegisterUser(ws *websocket.Conn, msg *websock.RegisterUserMessage) {
	// Add new user to database
	user := mdb.NewUser(msg.Username, msg.PublicKey)
	if err := s.Db.Insert(mdb.Users, user); err != nil {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Error registering user"})
		return
	}

	websock.Msg.Send(ws, &websock.Message{Type: websock.OK, Message: "User registered"})
}

// LoginUser authenticates a user using a randomly generated authentication token
// This token is encrypted with the public key of the username the client is trying to log in as
// The client is then expected to respond with the correct decrypted token
// TODO check if user is already logged in
func (s *Server) LoginUser(ws *websocket.Conn, username string) (*User, error) {

	// Create new user object
	newUser, encKey, err := NewUser(s.Db, username)
	if err != nil {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "User does not exist"})
		return nil, err
	}

	// Send auth challenge
	websock.Msg.Send(ws, &websock.Message{Type: websock.AuthChallenge, Message: encKey})

	// Receive auth challenge response
	res := new(websock.Message)
	if err := websock.Msg.Receive(ws, res); err != nil {
		log.Println(err)
		return nil, err
	}

	// Check that the received decrypted key matches the original auth key
	if !newUser.KeyMatches(res.Message.([]byte)) {
		websock.Msg.Send(ws, &websock.Message{Type: websock.Error, Message: "Invalid auth key"})
		return nil, errors.New("User provided invalid authentication key")
	}

	log.Printf("Client %s authenticated as user %s\n", ws.Request().RemoteAddr, newUser.Username)
	websock.Msg.Send(ws, &websock.Message{Type: websock.OK, Message: "Logged in"})
	return newUser, nil
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
