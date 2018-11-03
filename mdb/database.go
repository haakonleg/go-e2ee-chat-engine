package mdb

import (
	"log"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type Database struct {
	MongoURL string
	DBName   string

	session  *mgo.Session
	database *mgo.Database
}

func (d *Database) CreateConnection() error {
	session, err := mgo.Dial(d.MongoURL)
	if err != nil {
		return err
	}

	d.session = session
	d.database = session.DB(d.DBName)
	return nil
}

func (d *Database) InsertTest() {
	col := d.database.C("user")

	test := struct {
		ID   bson.ObjectId `bson:"id,omitempty"`
		Name string        `bson:"name"`
	}{ID: bson.NewObjectId(),
		Name: "Test1"}

	if err := col.Insert(test); err != nil {
		log.Fatal(err)
	}
}
