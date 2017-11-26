package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/akkgr/estiaServer/adapters"
)

const dbContextKey = adapters.DbContextKey
const userContextKey = adapters.UserContextKey

// Route describe a HTTP route
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

// Router ...
type Router interface {
	GetRoutes() []Route
}

// Controller struct
type Controller struct {
	dbName string
}

// NewController ...
func NewController(db string) *Controller {
	c := new(Controller)
	c.dbName = db
	return c
}

type dataResponse struct {
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
}

// SendJSON marshals v to a json struct and sends appropriate headers to w
func (c *Controller) SendJSON(v interface{}, w http.ResponseWriter) {
	json, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// GetContent of the request inside given struct
func (c *Controller) GetContent(v interface{}, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		return err
	}

	return nil
}

// HandleError write error on response and return false if there is no error
func (c *Controller) HandleError(err error, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return true
}
