package models

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
