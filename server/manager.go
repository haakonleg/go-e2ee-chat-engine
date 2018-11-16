package server

import (
	"fmt"
)

const creationSize = 3
const requestsSize = 10

// ChatRoomManager manages and runs all the chatrooms for the server
type ChatRoomManager struct {
	rooms map[string]*ChatRoom
	// Channel for handling requests to make chatrooms
	creation chan struct {
		name     string
		password string
		isHidden bool
		err      chan error
	}
	// Channel for requests to handle chatrooms
	requests chan struct {
		name string
		f    func(*ChatRoom)
		err  chan error
	}
	stop chan struct{}
}

// NewChatRoomManager makes a new chatroom manager
func NewChatRoomManager() (m ChatRoomManager) {
	m = ChatRoomManager{
		make(map[string]*ChatRoom),
		make(chan struct {
			name     string
			password string
			isHidden bool
			err      chan error
		}, creationSize),
		make(chan struct {
			name string
			f    func(*ChatRoom)
			err  chan error
		}, requestsSize),
		make(chan struct{}),
	}
	return
}

// Run processes creation for chatroom manager
func (m *ChatRoomManager) Run() {
	fmt.Println("Starting chatmanager")
	for {
		select {
		case req := <-m.creation:
			if _, ok := m.rooms[req.name]; ok {
				req.err <- fmt.Errorf("Chatroom with name '%s' already exists", req.name)
				continue
			}
			// Make and run new chatroom
			m.rooms[req.name] = NewChatRoom(req.name, req.password, req.isHidden)
			m.rooms[req.name].Run()
			req.err <- nil
		case req := <-m.requests:
			room, ok := m.rooms[req.name]
			if !ok {
				req.err <- fmt.Errorf("Chatroom with name '%s' does not exist", req.name)
				continue
			}
			req.f(room)
			req.err <- nil
		case <-m.stop:
			for _, room := range m.rooms {
				room.Stop()
			}
			fmt.Println("Shutting down chatmanager")
			return
		}
	}
}

// Stop the chatmanager event loop
func (m *ChatRoomManager) Stop() {
	m.stop <- struct{}{}
}

// NewRoom makes a new chatroom
func (m *ChatRoomManager) NewRoom(name, password string, isHidden bool) error {
	err := make(chan error, 1)
	m.creation <- struct {
		name     string
		password string
		isHidden bool
		err      chan error
	}{
		name,
		password,
		isHidden,
		err,
	}
	return <-err
}

// Interact with a chatroom by name and a callback function
func (m *ChatRoomManager) Interact(name string, f func(*ChatRoom)) error {
	err := make(chan error, 1)
	m.requests <- struct {
		name string
		f    func(*ChatRoom)
		err  chan error
	}{
		name,
		f,
		err,
	}
	return <-err
}
