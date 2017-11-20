package models

import (
	"gopkg.in/mgo.v2/bson"
)

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
