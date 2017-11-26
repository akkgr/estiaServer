package auth

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// HashPassword ...
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash ...
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// EnsureAdminUser ...
func EnsureAdminUser(s *mgo.Session, dbName string) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(dbName).C("users")
	var user User
	err := c.Find(bson.M{"username": "admin"}).One(&user)
	if err == nil {
		return
	}

	if err.Error() == "not found" {
		hash, err := HashPassword("admin")
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
