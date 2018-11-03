package websock

type MessageType int

const (
	RegisterUser MessageType = iota
)

type Message struct {
	Type    MessageType `json:"type"`
	Message interface{} `json:"message"`
}

type RegisterUserMessage struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}
