package main

import (
	"crypto/rand"
	"crypto/rsa"
	"io/ioutil"
	"log"

	"github.com/haakonleg/go-e2ee-chat-engine/util"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

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

	websock.Send(c.ws, &websock.Message{Type: websock.RegisterUser, Message: regUserMsg})

	_, err := c.wsReader.GetNext()
	if err != nil {
		c.gui.ShowDialog("Did not get a response from the server", nil)
		return
	}

	// Save private key to file
	savePrivKey(username, privKey)

	c.gui.ShowDialog("User created. You can now log in.", nil)
}

// Called when the user pressed the "login user" button
// TODO: Refactor the huge function
func (c *Client) loginUserHandler(server string, username string) {
	if !c.Connect(server) {
		return
	}

	// Read private key from file
	pem, err := ioutil.ReadFile(username + ".pem")
	if err != nil {
		c.gui.ShowDialog("Error reading privatekey.pem file", nil)
		return
	}

	privKey, err := util.UnmarshalPrivate(pem)
	if err != nil {
		c.gui.ShowDialog("Error parsing private key", nil)
		return
	}

	// Send log in request to server
	websock.Send(c.ws, &websock.Message{Type: websock.LoginUser, Message: username})

	// Receive auth challenge from server
	res, err := c.wsReader.GetNext()
	if err != nil {
		c.gui.ShowDialog(err.Error(), nil)
		return
	}
	log.Println(res)

	// Try to decrypt auth challenge
	decKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, res.Message.([]byte))
	if err != nil {
		c.gui.ShowDialog("Invalid private key", nil)
		return
	}
	log.Println(decKey)

	// Send decrypted auth key to server
	websock.Send(c.ws, &websock.Message{Type: websock.AuthChallengeResponse, Message: decKey})

	// Check response from server
	if res, err = c.wsReader.GetNext(); err != nil {
		c.gui.ShowDialog("Invalid private key", nil)
		return
	}
	log.Println(res)

	// Login success, show the chat rooms GUI
	c.privateKey = privKey
	c.authKey = decKey
	c.gui.ShowChatRoomGUI(c)
}

func (c *Client) createRoomHandler(name, password string, isHidden bool) {
	// Send request to create new chat room to server
	req := &websock.CreateChatRoomMessage{
		Name:     name,
		Password: password,
		IsHidden: isHidden}

	websock.Send(c.ws, &websock.Message{Type: websock.CreateChatRoom, Message: req})

	if _, err := c.wsReader.GetNext(); err != nil {
		c.gui.ShowDialog(err.Error(), nil)
	}
}

func (c *Client) getChatRooms() (*websock.GetChatRoomsResponseMessage, error) {
	// Send request for chat rooms
	websock.Send(c.ws, &websock.Message{Type: websock.GetChatRooms})

	// Get chat rooms response from server
	res, err := c.wsReader.GetNext()
	if err != nil {
		return nil, err
	}

	return res.Message.(*websock.GetChatRoomsResponseMessage), nil
}

func (c *Client) joinChatHandler(name, password string) {
	// Send request to join chat room
	req := &websock.JoinChatMessage{
		Name:     name,
		Password: password}

	websock.Send(c.ws, &websock.Message{Type: websock.JoinChat, Message: req})

	if _, err := c.wsReader.GetNext(); err != nil {
		c.gui.ShowDialog(err.Error(), nil)
		return
	}

	// Show the chat interface
	c.gui.ShowChatGUI(c)
}
