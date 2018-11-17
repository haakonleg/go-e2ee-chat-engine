package websock

// MessageType enum contains all possible websocket message types
type MessageType int

const (
	// Error means that en error occured
	Error MessageType = iota
	// OK means that the action was successfull
	OK

	// RegisterUser is sent when a client wants to register a new user
	RegisterUser
	// LoginUser is sent when a client wants to authenticate as a user
	LoginUser
	// AuthChallenge is sent by the server when an authentication challenge is initiated
	AuthChallenge
	// AuthChallengeResponse is sent by the client in resposne to an authentication challenge
	AuthChallengeResponse

	// CreateChatRoom is sent when a client wants to create a new chat room
	CreateChatRoom
	// GetChatRooms is sent when a client wants to retrieve a list of available chat rooms
	GetChatRooms
	// GetChatRoomsResponse is sent by the server in response to a GetChatRooms message
	GetChatRoomsResponse

	// JoinChat is sent when a clients wants to join a chat session
	JoinChat
	// ChatInfo is sent by the server when a client has joined a chat session
	ChatInfo
	// SendChat is sent when a client sends a chat message
	SendChat
	// ChatMessageReceived is sent by the server when another user in a chat room sends a chat message
	ChatMessageReceived
	// UserJoined is sent by the server when a user joins a chat room the client is in
	UserJoined
	// UserLeft is sent by the server when a user leaves a chat room the client is in
	UserLeft
	// LeaveChat is sent when a client wants to leave a chat room
	LeaveChat

	// Ping is a keepalive message sent by the server
	Ping
	// Pong is sent by the client in response to a Ping message
	Pong
)

// Message is the "base" message which is used for all websocket messages
// Type contains the type of the message (one of the MessageType enums)
// Message contains the actual content of the message, which can be a string, byte slice, a struct, or nil.
type Message struct {
	Type    MessageType
	Message interface{}
}

// RegisterUserMessage is the message sent by a client to request user registration
type RegisterUserMessage struct {
	Username  string
	PublicKey []byte
}

// CreateChatRoomMessage is the message sent by a client to request creation of a new chat room
type CreateChatRoomMessage struct {
	Name     string
	Password string
	IsHidden bool
}

// GetChatRoomsResponseMessage is sent by the server in response to a GetChatRooms request
type GetChatRoomsResponseMessage struct {
	TotalConnected int
	Rooms          []Room
}

// Room is used in GetChatRoomsResponse
type Room struct {
	Name        string
	HasPassword bool
	OnlineUsers int
}

// JoinChatMessage is the message sent by a client to request to join a chat room
type JoinChatMessage struct {
	Name     string
	Password string
}

// ChatInfoMessage is the message sent by the server to a client who joined a chat room
type ChatInfoMessage struct {
	Name       string
	MyUsername string
	Users      []User
	Messages   []*ChatMessage
}

// User is used in ChatInfoMessage, and by the server when notifying a client about a new connected user
type User struct {
	Username  string
	PublicKey []byte
}

// ChatMessage is used in ChatInfoMessage, and by the server when notifying a client about a new chat message
type ChatMessage struct {
	Sender    string
	Timestamp int64
	Message   []byte
}

// SendChatMessage is the message sent by the client to the server when a new chat message is sent.
// The map EncryptedContent contains the message content encrypted by every recipients public key
type SendChatMessage struct {
	EncryptedContent map[string][]byte
}
