package mdb

import (
	"github.com/globalsign/mgo/bson"
)

type Message struct {
	ID               bson.ObjectId `bson:"_id"`
	Timestamp        int64         `bson:"timestamp"`
	Sender           string        `bson:"sender"`
	EncryptedMessage []byte        `bson:"encrypted_message"`
}

type Chat struct {
	ID        bson.ObjectId `bson:"_id"`
	Timestamp int64         `bson:"timestamp"`
	Name      string        `bson:"name"`
	Messages  []Message     `bson:"messages"`

	// This field is not used in the database
	Users []string `bson:"-"`
}

// NewChat creates a new instance of the Chat object
func NewChat(name string) *Chat {
	return &Chat{
		ID:       bson.NewObjectId(),
		Name:     name,
		Users:    make([]string, 0),
		Messages: make([]Message, 0)}
}
