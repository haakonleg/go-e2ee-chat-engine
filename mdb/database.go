package mdb

import (
	"errors"
	"log"

	"github.com/globalsign/mgo"
)

// DatabaseCollection is used to refer to allowed database collections in functions
type DatabaseCollection int

const (
	// Users is the collection containing users
	Users DatabaseCollection = iota
	ChatRooms
	Messages
)

func (c DatabaseCollection) String() string {
	switch c {
	case Users:
		return "users"
	case ChatRooms:
		return "chat_rooms"
	case Messages:
		return "messages"
	}
	return ""
}

type Database struct {
	dbName  string
	session *mgo.Session
}

// CreateConnection creates a new connection to the database
func CreateConnection(mongoURL, dbName string) (*Database, error) {
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return nil, err
	}

	db := &Database{
		dbName:  dbName,
		session: session}

	db.MakeIndexes()
	return db, nil
}

// DeleteAll removes all data inside all collections, but not the information about the
// collections themselves
func (db *Database) DeleteAll() {
	var err error

	// Indexes for users
	c := db.session.DB(db.dbName).C(Users.String())
	err = c.DropAllIndexes()
	if err != nil {
		log.Printf("Unable to drop indexes of %s: %s\n", Users.String(), err)
	}

	// Indexes for chat rooms
	c = db.session.DB(db.dbName).C(ChatRooms.String())
	err = c.DropAllIndexes()
	if err != nil {
		log.Printf("Unable to drop indexes of %s: %s\n", ChatRooms.String(), err)
	}

	// Indexes for messages
	c = db.session.DB(db.dbName).C(Messages.String())
	err = c.DropAllIndexes()
	if err != nil {
		log.Printf("Unable to drop indexes of %s: %s\n", Messages.String(), err)
	}
}

// MakeIndexes creates necessary indexes and unique constraints for keys in the database
func (db *Database) MakeIndexes() {
	// Indexes for users
	c := db.session.DB(db.dbName).C(Users.String())
	c.EnsureIndex(mgo.Index{
		Key:    []string{"username"},
		Unique: true})

	// Indexes for chat rooms
	c = db.session.DB(db.dbName).C(ChatRooms.String())
	c.EnsureIndex(mgo.Index{
		Key:    []string{"name"},
		Unique: true})

	// Indexes for messages
	c = db.session.DB(db.dbName).C(Messages.String())
	c.EnsureIndex(mgo.Index{
		Key:    []string{"chat_name"},
		Unique: false})
}

// Insert inserts one or more objects into the database, creates a temporary copy of the session for better concurrency performance
func (db *Database) Insert(collection DatabaseCollection, objects ...interface{}) error {
	sessionCpy := db.session.Copy()
	defer sessionCpy.Close()

	col := sessionCpy.DB(db.dbName).C(collection.String())
	if err := col.Insert(objects...); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (db *Database) FindAll(collection DatabaseCollection, query interface{}, selector interface{}, result interface{}) error {
	sessionCpy := db.session.Copy()
	defer sessionCpy.Close()

	q := sessionCpy.DB(db.dbName).C(collection.String()).Find(query)
	if selector != nil {
		q = q.Select(selector)
	}

	if err := q.All(result); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (db *Database) FindOne(collection DatabaseCollection, query interface{}, selector interface{}, result interface{}) error {
	sessionCpy := db.session.Copy()
	defer sessionCpy.Close()

	q := sessionCpy.DB(db.dbName).C(collection.String()).Find(query)
	if selector != nil {
		q = q.Select(selector)
	}

	if cnt, err := q.Count(); err != nil {
		log.Println(err)
		return err
	} else if cnt == 0 {
		return errors.New("Got 0 results")
	}

	if err := q.One(result); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
