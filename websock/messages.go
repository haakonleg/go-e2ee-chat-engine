package websock

// MessageType enum contains all possible websocket message types
type MessageType int

// MessageFormat enum contains all possible websocket message formats
type MessageFormat int

const (
	Error MessageType = iota
	MessageOK

	// User/auth related
	RegisterUser
	LoginUser
	AuthChallenge
	AuthChallengeResponse

	// Chat related
	CreateChatRoom
	GetChatRooms
	GetChatRoomsResponse
	JoinChat
	ChatInfo
	SendChat
	ChatMessageReceived

	JSON MessageFormat = iota
	String
	Bytes
	Nil
)

// Message is the "base" message which is used for all websocket messages
// Type contains the type of the message (one of the MessageType enums)
// Message contains marshalled JSON, a string, a byte array, or Nil (no message)
type Message struct {
	Type    MessageType `json:"type"`
	Message []byte      `json:"message"`
}

// RegisterUserMessage is the message sent by a client to request user registration
type RegisterUserMessage struct {
	Username  string `json:"username"`
	PublicKey []byte `json:"public_key"`
}

// CreateChatRoomMessage is the message sent by a client to request creation of a new chat room
type CreateChatRoomMessage struct {
	Name    string `json:"name"`
	AuthKey []byte `json:"auth_key"`
}

// GetChatRoomsResponseMessage is sent by the server in response to a GetChatRooms request
type GetChatRoomsResponseMessage struct {
	Rooms []Room `json:"rooms"`
}

// Room is used in GetChatRoomsResponse
type Room struct {
	Name        string `json:"name"`
	OnlineUsers int    `json:"online_users"`
}

// JoinChatMessage is the message sent by a client to request to join a chat room
type JoinChatMessage struct {
	Name    string `json:"name"`
	AuthKey []byte `json:"auth_key"`
}

// ChatInfoMessage is the message sent by the server to a client who joined a chat room
type ChatInfoMessage struct {
	Name     string         `json:"name"`
	Users    []User         `json:"users"`
	Messages []*ChatMessage `json:"messages"`
}

// User is used in ChatInfoMessage
type User struct {
	Username  string `json:"username"`
	PublicKey []byte `json:"public_key"`
}

// ChatMessage is used in ChatInfoMessage, and by the server when notifying a client about a new chat message
type ChatMessage struct {
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
	Message   []byte `json:"message"`
}

// SendChatMessage is the message sent by the client to the server when a new chat message is sent.
// The map EncryptedContent contains the message content encrypted by every recipients public key
type SendChatMessage struct {
	EncryptedContent map[string][]byte `json:"encrypted_content"`
	AuthKey          []byte            `json:"auth_key"`
}
