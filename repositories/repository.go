package repositories

import (
	mgo "gopkg.in/mgo.v2"
)

// Context database context
type Context struct {
	dbName string
}

// NewContext creates a new database context
func NewContext(db string) *Context {
	c := new(Context)
	c.dbName = db
	return c
}

// EnsureIndex ...
func (ctx *Context) EnsureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB(ctx.dbName).C("users")

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
