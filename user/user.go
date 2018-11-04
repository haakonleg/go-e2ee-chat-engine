package user

import (
	"crypto/rsa"

	"github.com/globalsign/mgo/bson"
	"github.com/haakonleg/go-e2ee-chat-engine/mdb"
)

type User struct {
	Username  string
	AuthKey   []byte
	EncKey    []byte
	PublicKey *rsa.PublicKey
}

// RegisterUser registers a new user in the database
func RegisterUser(db *mdb.Database, username string, publicKey []byte) error {
	user := mdb.NewUser(username, publicKey)

	if err := db.Insert(mdb.Users, []interface{}{user}); err != nil {
		return err
	}
	return nil
}

// AuthChallenge generates an authentication challenge based on the users public key.
// A random string is generated and encrypted with the users public key, the user
// is expected to send in response the decrypted string
func AuthChallenge(db *mdb.Database, username string) (*User, error) {
	// Retrieve user from DB
	query := bson.M{"username": username}

	user := new(mdb.User)
	if err := db.FindOne(mdb.Users, query, user); err != nil {
		return nil, err
	}

	// Unmarshal public key
	pubKey, err := UnmarshalPublic(user.PublicKey)
	if err != nil {
		return nil, err
	}

	// Generate auth challenge
	encKey, authKey := genAuthChallenge(pubKey)

	return &User{
		Username:  username,
		AuthKey:   authKey,
		EncKey:    encKey,
		PublicKey: pubKey}, nil
}
