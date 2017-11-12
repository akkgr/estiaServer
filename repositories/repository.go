package repositories

import (
	"log"

	"github.com/akkgr/estiaServer/models"
	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MySigningKey ...
var MySigningKey = []byte("TooSlowTooLate4u.")

// DbName ...
var DbName = "estiag"

// EnsureIndex ...
func EnsureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(DbName).C("users")

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

// EnsureAdminUser ...
func EnsureAdminUser(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(DbName).C("users")
	var user models.User
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
