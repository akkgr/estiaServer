package models

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
