# go-e2ee-chat-engine

[![Build Status](https://travis-ci.org/haakonleg/go-e2ee-chat-engine.svg?branch=master)](https://travis-ci.org/haakonleg/go-e2ee-chat-engine)
[![Go Report Card](https://goreportcard.com/badge/github.com/haakonleg/go-e2ee-chat-engine)](https://goreportcard.com/report/github.com/haakonleg/go-e2ee-chat-engine)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/haakonleg/go-e2ee-chat-engine)
[![Release](https://img.shields.io/github/release/haakonleg/go-e2ee-chat-engine.svg)](https://github.com/haakonleg/go-e2ee-chat-engine/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/haakonleg/go-e2ee-chat-engine/badge.svg?branch=master)](https://coveralls.io/github/haakonleg/go-e2ee-chat-engine?branch=master)

Golang chat engine backend with end-to-end encryption.
Project in the course IMT2681 Cloud Technologies, assignment 3.

## Project Description

The project idea is to develop a chat engine utilizing end-to-end encryption with RSA2048. Any registered user can create a chat room (with optional password), and others can join. Users in a chat room each have a private key (secret) and public key, and each users public key are stored on the server. When a user sends a message, he encrypts the message with each recipients public key. The backend server must keep track of connected users, public keys, chat rooms, encrypted messages. The communication between client and server will happen through a websocket.

The server will be deployed on Heroku as a Docker image. A simple (command line) demonstration client will be created.

## Todo

- Add ability to set a password for a chat room.
- At the moment, a user cannot see messages that is sent when he is not in a chat room the moment it is sent (because clients in chat rooms are not tracked in the database, but in-memory on the server). Fix this.
- Allow users to be part of multiple chat rooms (see above).
- Add a server setting to purge old chat messages after a certain date (to avoid massive amounts of old messages)
- Implement concept of a chat room admin/owner (and add ability to delete/rename chat room, kick/ban users)
- ~~Allow user to leave a chat in the client app~~
- ~~The chat room list in the client is not good (when it refreshes every 2 seconds the user selection is lost). To fix this do not clear the entire list when it is refreshed, but add only new chat rooms to the list on refresh.~~
- ~~Prevent users from registering a user with a empty username++~~ (@barskern)
- Add validation of messages on the server side
- ~~The server code is probably not thread-safe (ConnectedClients map in server.go), we need to redisign the way we access the clients and currently connected users. Probably need to find a way to not have to use mutexes directly, but create some kind of abstraction to access the connected clients.~~
- Seperate validation from the server code to another file/package, and ensure that validation is being done server side for usernames, chat room names, chat messages etc...
- Ensure that chat rooms and messages are being fetched from the database in a preffered order. For example maybe chat rooms should be listed in descending order according to number of users, then the timestamp etc... And chat messages must be listed according to the timestamp. This is currently not ensured in the server code.
