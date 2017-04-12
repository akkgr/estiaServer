package main

import (
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(db).C("users")

	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func ensureAdminUser(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(db).C("users")
	var user user
	err := c.Find(bson.M{"username": "admin"}).One(&user)
	if err == nil {
		return
	}

	if err.Error() == "not found" {
		hash, err := hashPassword("admin")
		if err != nil {
			log.Fatal(err)
		}
		user.Username = "admin"
		user.Password = hash
		err = c.Insert(user)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
