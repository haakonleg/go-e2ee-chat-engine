# go-e2ee-chat-engine
Golang chat engine backend with end-to-end encryption.
Project in the course IMT2681 Cloud Technologies, assignment 3.

## Project Description
The project idea is to develop a chat engine utilizing end-to-end encryption with RSA2048. Any registered user can create a chat room (with optional password), and others can join. Users in a chat room each have a private key (secret) and public key, and each users public key are stored on the server. When a user sends a message, he encrypts the message with each recipients public key. The backend server must keep track of connected users, public keys, chat rooms, encrypted messages. The communication between client and server will happen through a websocket.

The server will be deployed on OpenStack as a Docker image. A simple (command line) demonstration client will be created.

## Libraries to use
https://godoc.org/golang.org/x/net/websocket

https://golang.org/pkg/crypto/rsa/

https://github.com/globalsign/mgo
