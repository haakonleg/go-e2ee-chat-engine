package mdb

import (
	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
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

func NewMessage(chatName, sender string) *Message {
	return &Message{
		ID:             bson.NewObjectId(),
		ChatName:       chatName,
		Timestamp:      util.NowMillis(),
		Sender:         sender,
		MessageContent: make([]MessageContent, 0)}
}
