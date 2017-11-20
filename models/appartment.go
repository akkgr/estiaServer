package models

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
