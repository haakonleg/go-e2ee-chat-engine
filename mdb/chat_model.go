package mdb

import (
	"crypto/sha256"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/util"
)

type Chat struct {
	ID           bson.ObjectId `bson:"_id"`
	Timestamp    int64         `bson:"timestamp"`
	Name         string        `bson:"name"`
	PasswordHash []byte        `bson:"password_hash"`
	IsHidden     bool          `bson:"is_hidden"`
}

// ValidPassword compares the checksum of a plaintext password to the checksum
// stored in the Chat object.
func (c *Chat) ValidPassword(password string) bool {
	if len(c.PasswordHash) == 0 {
		return true
	}

	passwordHash := sha256.Sum256([]byte(password))
	for i := 0; i < sha256.Size; i++ {
		if c.PasswordHash[i] != passwordHash[i] {
			return false
		}
	}
	return true
}

// NewChat creates a new instance of the Chat object. It takes a plaintext password
// as input, and returns a new Chat object containing the hashed checksum of that password.
func NewChat(name, password string, isHidden bool) *Chat {
	passwordHash := make([]byte, 0)
	if len(password) > 1 {
		hash := sha256.Sum256([]byte(password))
		passwordHash = hash[:]
	}

	return &Chat{
		ID:           bson.NewObjectId(),
		Timestamp:    util.NowMillis(),
		Name:         name,
		IsHidden:     isHidden,
		PasswordHash: passwordHash}
}
