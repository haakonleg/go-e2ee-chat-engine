package mdb

import (
	"github.com/globalsign/mgo/bson"
)

// Message is the model of chat messages stored in the database
type Message struct {
	ID             bson.ObjectId    `bson:"_id"`
	ChatName       string           `bson:"chat_name"`
	Timestamp      int64            `bson:"timestamp"`
	Sender         string           `bson:"sender"`
	MessageContent []MessageContent `bson:"message_content"`
}

// MessageContent contains the ciphertext of a chat message addressed to a specific user
// There should be an entry for each recipient in the chat room when the chat message was sent.
type MessageContent struct {
	Recipient string `bson:"recipient"`
	Content   []byte `bson:"content"`
}

// NewMessage creates a new instance of the Message object
func NewMessage(chatName string, timestamp int64, sender string) *Message {
	return &Message{
		ID:             bson.NewObjectId(),
		ChatName:       chatName,
		Timestamp:      timestamp,
		Sender:         sender,
		MessageContent: make([]MessageContent, 0)}
}
