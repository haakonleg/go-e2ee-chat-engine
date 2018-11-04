package websock

type MessageType int
type MessageFormat int

const (
	Error MessageType = iota
	MessageOK
	RegisterUser
	LoginUser
	AuthChallenge
	ChallengeResponse

	JSON MessageFormat = iota
	String
	Bytes
)

type Message struct {
	Type    MessageType `json:"type"`
	Message []byte      `json:"message"`
}

type RegisterUserMessage struct {
	Username  string `json:"username"`
	PublicKey []byte `json:"public_key"`
}
