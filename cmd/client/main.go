package main

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/haakonleg/go-e2ee-chat-engine/user"
	"golang.org/x/net/websocket"

	"github.com/haakonleg/go-e2ee-chat-engine/websock"
)

const (
	privKeyFile = "privatekey.pem"
)

var socket *websocket.Conn

func savePrivKey(privKey *rsa.PrivateKey) {
	pem := user.MarshalPrivate(privKey)
	if err := ioutil.WriteFile(privKeyFile, pem, 0644); err != nil {
		log.Fatal(err)
	}
}

// Called when user pressed the "create user" button
func createUserHandler(gui *LoginGUI, server string, username string) error {
	// Generate new key pair
	privKey, pubKey := user.GenKeyPair()

	// Send a request to register the user
	regUserMsg := &websock.RegisterUserMessage{
		Username:  username,
		PublicKey: user.MarshalPublic(pubKey)}

	if err := websock.SendMessage(socket, websock.RegisterUser, regUserMsg, websock.JSON); err != nil {
		return err
	}

	res, err := websock.GetResponse(socket)
	if err != nil {
		return errors.New("Did not get response from server")
	}

	if res.Type == websock.MessageOK {
		// Save private key to file
		savePrivKey(privKey)
		gui.app.Stop()
	}

	return nil
}

// Called when the user pressed the "login user" button
func loginUserHandler(gui *LoginGUI, server string, username string) error {
	// Get private key
	pem, err := ioutil.ReadFile(privKeyFile)
	if err != nil {
		return errors.New("Error reading privatekey.pem file")
	}

	privKey, err := user.UnmarshalPrivate(pem)
	if err != nil {
		return errors.New("Error parsing private key")
	}

	// Recieve auth challenge from server
	if err := websock.SendMessage(socket, websock.LoginUser, username, websock.String); err != nil {
		return err
	}
	res, err := websock.GetResponse(socket)
	if err != nil {
		return errors.New("Error receiving auth challenge from server")
	}

	// Decrypt auth challenge
	decKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, res.Message)
	if err != nil {
		return errors.New("Error decrypting auth key")
	}

	// Send decrypted auth key to server
	if err := websock.SendMessage(socket, websock.ChallengeResponse, decKey, websock.Bytes); err != nil {
		return err
	}

	// Check response from server
	res, err = websock.GetResponse(socket)
	if res.Type == websock.MessageOK {
		return errors.New("Logged in!")
	} else {
		return errors.New("Invalid private key")
	}

	return nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("Usage: %s {server}\n", args[0])
		os.Exit(1)
	}
	server := args[1]

	// Try to connect to websocket
	ws, err := NewClient(server)
	if err != nil {
		log.Fatal("Unable to connect to server")
	}
	socket = ws

	loginGUI := &LoginGUI{
		DefaultServerText: "ws://localhost:5000",
		CreateUserHandler: createUserHandler,
		LoginUserHandler:  loginUserHandler}

	loginGUI.Create()
	loginGUI.Show()
}
