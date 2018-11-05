package websock

// MessageType enum contains all possible websocket message types
type MessageType int

// MessageFormat enum contains all possible websocket message formats
type MessageFormat int

const (
	Error MessageType = iota
	MessageOK
	RegisterUser
	LoginUser
	AuthChallenge
	ChallengeResponse
	CreateChatRoom
	GetChatRooms
	ChatRoomsResponse

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

// GetChatRoomsResponse is sent by the server in response to a GetChatRooms request
type GetChatRoomsResponse struct {
	Rooms []Room `json:"rooms"`
}

// Room is used in GetChatRoomsResponse
type Room struct {
	Name        string `json:"name"`
	OnlineUsers int    `json:"online_users"`
}
