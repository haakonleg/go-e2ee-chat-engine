package mdb

import (
	"github.com/globalsign/mgo/bson"
)

type Message struct {
	ID             bson.ObjectId    `bson:"_id"`
	ChatName       string           `bson:"chat_name"`
	Timestamp      int64            `bson:"timestamp"`
	Sender         string           `bson:"sender"`
	MessageContent []MessageContent `bson:"message_content"`
}

type MessageContent struct {
	Recipient string `bson:"recipient"`
	Content   []byte `bson:"content"`
}

func NewMessage(chatName string, timestamp int64, sender string) *Message {
	return &Message{
		ID:             bson.NewObjectId(),
		ChatName:       chatName,
		Timestamp:      timestamp,
		Sender:         sender,
		MessageContent: make([]MessageContent, 0)}
}
