package websock

// MessageType enum contains all possible websocket message types
type MessageType int

const (
	Error MessageType = iota
	OK

	// User/auth related
	RegisterUser
	LoginUser
	AuthChallenge
	AuthChallengeResponse

	// Chat related
	CreateChatRoom
	GetChatRooms
	GetChatRoomsResponse

	// Messages used for a chat session
	JoinChat
	ChatInfo
	SendChat
	ChatMessageReceived
	UserJoined
	UserLeft
	LeaveChat

	// Keepalive messages
	Ping
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
