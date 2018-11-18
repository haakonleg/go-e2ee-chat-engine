package server

import (
	"fmt"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
	"github.com/haakonleg/go-e2ee-chat-engine/websock"
	"sync"
)

const creationSize = 3
const requestsSize = 10

// ChatRoomManager manages and runs all the chatrooms for the server
type ChatRoomManager struct {
	sync.RWMutex
	rooms map[string]*ChatRoom
	// DB is a connection to the database
	db *mdb.Database
}

// NewChatRoomManager makes a new chatroom manager
func NewChatRoomManager(db *mdb.Database) (m *ChatRoomManager) {
	m = &ChatRoomManager{
		sync.RWMutex{},
		make(map[string]*ChatRoom),
		db,
	}
	return
}

// NewRoom makes a new chatroom
func (m *ChatRoomManager) NewRoom(name, password string, isHidden bool) error {
	m.RLock()
	if _, ok := m.rooms[name]; ok {
		m.RUnlock()
		return fmt.Errorf("Chatroom with name (%s) already exists", name)
	}
	m.RUnlock()

	room, err := NewChatRoom(m.db, name, password, isHidden)
	if err != nil {
		return err
	}

	m.Lock()
	defer m.Unlock()
	m.rooms[name] = room
	room.Run()

	return nil
}

// Interact with a chatroom by name and a callback function
func (m *ChatRoomManager) Interact(name string, f func(*ChatRoom)) error {
	m.RLock()
	defer m.RUnlock()
	room, ok := m.rooms[name]
	if !ok {
		return fmt.Errorf("Chatroom with name (%s) does not exist", name)
	}

	f(room)
	return nil
}

// GetRooms returns information about all chatrooms
func (m *ChatRoomManager) GetRooms() (rooms []websock.Room) {
	m.RLock()
	defer m.RUnlock()
	for roomname, room := range m.rooms {
		rooms = append(rooms, websock.Room{
			Name:        roomname,
			HasPassword: room.HasPassword(),
			OnlineUsers: room.UserCount(),
		})
	}
	return
}
