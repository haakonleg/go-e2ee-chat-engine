package mdb

import (
	"github.com/globalsign/mgo/bson"
)

// User is the model of a user stored in the database
type User struct {
	ID        bson.ObjectId `bson:"_id"`
	Username  string        `bson:"username"`
	PublicKey []byte        `bson:"public_key"`
}

// NewUser creates a new instance of the user object
func NewUser(username string, publicKey []byte) *User {
	return &User{
		ID:        bson.NewObjectId(),
		Username:  username,
		PublicKey: publicKey}
}
