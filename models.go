package main

type geoLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type address struct {
	Area         string      `json:"area"`
	Street       string      `json:"street"`
	StreetNumber string      `json:"streetNumber"`
	PostalCode   string      `json:"postalCode"`
	Country      string      `json:"country"`
	Location     geoLocation `json:"location"`
}

type person struct {
	Lastname  string  `json:"lastname"`
	Firstname string  `json:"firstname"`
	Address   address `json:"address"`
	Home      string  `json:"home"`
	Work      string  `json:"work"`
	Mobile    string  `json:"mobile"`
	Fax       string  `json:"fax"`
	Other     string  `json:"other"`
	Email     string  `json:"email"`
	Ibank     string  `json:"ibank"`
}

type appartment struct {
	Title    string   `json:"title"`
	Position int32    `json:"position"`
	Resident person   `json:"resident"`
	Owner    person   `json:"owner"`
	Common   int64    `json:"common"`
	Elevetor int64    `json:"elevetor"`
	Heat     int64    `json:"heat"`
	Ei       int64    `json:"ei"`
	Fi       int64    `json:"fi"`
	Owners   int64    `json:"owners"`
	Other    int64    `json:"other"`
	Counters []string `json:"counters"`
}

type building struct {
	Address     address      `json:"address"`
	Oil         int64        `json:"oil"`
	Fund        int64        `json:"fund"`
	Closed      int64        `json:"closed"`
	Active      bool         `json:"active"`
	Managment   bool         `json:"managment"`
	Appartments []appartment `json:"appartments"`
}
