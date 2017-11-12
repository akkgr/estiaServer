package models

import (
	"gopkg.in/mgo.v2/bson"
)

// JwtToken helper type
type JwtToken struct {
	Token string `json:"token"`
}

// DataResponse helper type
type DataResponse struct {
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
}

// User Model
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Address Model
type Address struct {
	Area         string  `json:"area"`
	Street       string  `json:"street"`
	StreetNumber string  `json:"streetNumber"`
	PostalCode   string  `json:"postalCode"`
	Country      string  `json:"country"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
}

// Person Model
type Person struct {
	Lastname  string  `json:"lastname"`
	Firstname string  `json:"firstname"`
	Address   Address `json:"address"`
	Home      string  `json:"home"`
	Work      string  `json:"work"`
	Mobile    string  `json:"mobile"`
	Fax       string  `json:"fax"`
	Other     string  `json:"other"`
	Email     string  `json:"email"`
	Ibank     string  `json:"ibank"`
}

// Appartment Model
type Appartment struct {
	Title    string   `json:"title"`
	Position int32    `json:"position"`
	Resident Person   `json:"resident"`
	Owner    Person   `json:"owner"`
	Common   int64    `json:"common"`
	Elevetor int64    `json:"elevetor"`
	Heat     int64    `json:"heat"`
	Ei       int64    `json:"ei"`
	Fi       int64    `json:"fi"`
	Owners   int64    `json:"owners"`
	Other    int64    `json:"other"`
	Counters []string `json:"counters"`
}

// Building Model
type Building struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Address     Address       `json:"address"`
	Oil         int64         `json:"oil"`
	Fund        int64         `json:"fund"`
	Closed      int64         `json:"closed"`
	Active      bool          `json:"active"`
	Managment   bool          `json:"managment"`
	Appartments []Appartment  `json:"appartments"`
	Username    string        `json:"username"`
}
