package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

func savePrivKey(privKey *rsa.PrivateKey) {
	pem := util.MarshalPrivate(privKey)
	if err := ioutil.WriteFile(privKeyFile, pem, 0644); err != nil {
		log.Fatal(err)
	}
}

// Called when user pressed the "create user" button
func (c *Client) createUserHandler(server string, username string) {
	if !c.Connect(server) {
		return
	}

	// Generate new key pair
	privKey, pubKey := util.GenKeyPair()

	// Send a request to register the user
	regUserMsg := &websock.RegisterUserMessage{
		Username:  username,
		PublicKey: util.MarshalPublic(pubKey)}

	websock.SendMessage(c.sock, websock.RegisterUser, regUserMsg, websock.JSON)

	res, err := websock.GetResponse(c.sock)
	if err != nil {
		c.ShowDialog("Did not get a response from the server")
		return
	}

	if res.Type == websock.MessageOK {
		// Save private key to file
		savePrivKey(privKey)
	}

	c.ShowDialog("User created. You can now log in.")
}

// Called when the user pressed the "login user" button
// TODO: Refactor the huge function
func (c *Client) loginUserHandler(server string, username string) {
	if !c.Connect(server) {
		return
	}

	// Read private key from file
	pem, err := ioutil.ReadFile(privKeyFile)
	if err != nil {
		c.ShowDialog("Error reading privatekey.pem file")
		return
	}

	privKey, err := util.UnmarshalPrivate(pem)
	if err != nil {
		c.ShowDialog("Error parsing private key")
		return
	}

	// Send log in request to server
	websock.SendMessage(c.sock, websock.LoginUser, username, websock.String)

	// Recieve auth challenge from server
	res, err := websock.GetResponse(c.sock)
	if err != nil {
		c.ShowDialog("Error receiving auth challenge from server")
		return
	} else if res.Type == websock.Error {
		c.ShowDialog(string(res.Message))
		return
	}

	// Try to decrypt auth challenge
	decKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, res.Message)
	if err != nil {
		c.ShowDialog("Invalid private key")
		return
	}

	// Send decrypted auth key to server
	websock.SendMessage(c.sock, websock.ChallengeResponse, decKey, websock.Bytes)

	// Check response from server
	if res, err = websock.GetResponse(c.sock); err != nil || res.Type != websock.MessageOK {
		c.ShowDialog("Invalid private key")
		return
	}

	// Login success, show the chat rooms GUI
	c.authKey = decKey
	c.ShowChatRoomGUI()
}

func (c *Client) createNewChatRoomHandler(name string) {
	// Send request to create new chat room to server
	req := &websock.CreateChatRoomMessage{
		Name:    name,
		AuthKey: c.authKey}

	websock.SendMessage(c.sock, websock.CreateChatRoom, req, websock.JSON)
}

func (c *Client) getChatRooms() *websock.GetChatRoomsResponse {
	// Send request for chat rooms
	websock.SendMessage(c.sock, websock.GetChatRooms, nil, websock.Nil)

	// Get chat rooms response from server
	res, err := websock.GetResponse(c.sock)
	if err != nil || res.Type != websock.ChatRoomsResponse {
		c.ShowDialog("Error getting chat rooms from server")
		return nil
	}

	// Unmarshal response
	chatRoomsResponse := new(websock.GetChatRoomsResponse)
	if err := json.Unmarshal(res.Message, chatRoomsResponse); err != nil {
		c.ShowDialog("Error parsing chat rooms response")
		return nil
	}
	return chatRoomsResponse
}
