package mdb

import (
	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
)

type Chat struct {
	ID        bson.ObjectId `bson:"_id"`
	Timestamp int64         `bson:"timestamp"`
	Name      string        `bson:"name"`
}

// NewChat creates a new instance of the Chat object
func NewChat(name string) *Chat {
	return &Chat{
		ID:        bson.NewObjectId(),
		Timestamp: util.NowMillis(),
		Name:      name}
}
