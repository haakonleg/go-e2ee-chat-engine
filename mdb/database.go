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
)

func (c DatabaseCollection) String() string {
	switch c {
	case Users:
		return "users"
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

	return &Database{
		dbName:  dbName,
		session: session}, nil
}

// Insert inserts one or more objects into the database, creates a temporary copy of the session for better concurrency performance
func (db *Database) Insert(collection DatabaseCollection, objects []interface{}) error {
	sessionCpy := db.session.Copy()
	defer sessionCpy.Close()

	col := sessionCpy.DB(db.dbName).C(collection.String())
	if err := col.Insert(objects...); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (db *Database) FindOne(collection DatabaseCollection, query interface{}, result interface{}) error {
	sessionCpy := db.session.Copy()
	defer sessionCpy.Close()

	q := sessionCpy.DB(db.dbName).C(collection.String()).Find(query)

	if cnt, err := q.Count(); err != nil {
		return err
	} else if cnt == 0 {
		return errors.New("Got 0 results")
	}

	if err := q.One(result); err != nil {
		return err
	}

	return nil
}
