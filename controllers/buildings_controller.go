package controllers

import (
	"net/http"
	"strconv"

	"github.com/akkgr/estiaServer/models"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	tableName   = "buildings"
	routePrefix = "buildings"
)

// BuildingsController struct
type BuildingsController struct {
	Controller
}

// GetRoutes ...
func (b BuildingsController) GetRoutes() []Route {
	return []Route{
		Route{
			Method:  http.MethodGet,
			Path:    "/" + routePrefix + "/{offset}/{limit}",
			Handler: b.list,
		},
		Route{
			Method:  http.MethodGet,
			Path:    "/" + routePrefix + "/{id}",
			Handler: b.get,
		},
		Route{
			Method:  http.MethodPost,
			Path:    "/" + routePrefix,
			Handler: b.create,
		},
		Route{
			Method:  http.MethodPut,
			Path:    "/" + routePrefix + "{id}",
			Handler: b.update,
		},
		Route{
			Method:  http.MethodDelete,
			Path:    "/" + routePrefix + "{id}",
			Handler: b.delete,
		},
	}
}

func (b BuildingsController) list(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	vars := mux.Vars(r)
	offset, _ := strconv.Atoi(vars["offset"])
	limit, _ := strconv.Atoi(vars["limit"])

	c := session.DB(b.dbName).C(tableName)

	var data []models.Building
	count, err := c.Find(bson.M{}).Count()
	if b.HandleError(err, w) {
		return
	}

	err = c.Find(bson.M{}).Sort(
		"address.street",
		"address.streetNumber",
		"address.area",
		"address.country").Skip(offset).Limit(limit).All(&data)
	if b.HandleError(err, w) {
		return
	}

	var resp dataResponse
	resp.Data = data
	resp.Count = count
	b.SendJSON(resp, w)
}

func (b BuildingsController) get(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	vars := mux.Vars(r)
	id := vars["id"]

	var data models.Building

	if id == "0" {
		b.SendJSON(data, w)
		return
	}

	c := session.DB(b.dbName).C(tableName)
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&data)
	if b.HandleError(err, w) {
		return
	}

	b.SendJSON(data, w)
}

func (b BuildingsController) create(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	data := models.Building{}
	err := b.GetContent(&data, r)
	if b.HandleError(err, w) {
		return
	}

	c := session.DB(b.dbName).C(tableName)

	err = c.Insert(data)
	if b.HandleError(err, w) {
		return
	}

	b.SendJSON(data, w)
}

func (b BuildingsController) update(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)
	username := r.Context().Value(userContextKey).(string)

	vars := mux.Vars(r)
	id := vars["id"]

	data := models.Building{}
	err := b.GetContent(&data, r)
	if b.HandleError(err, w) {
		return
	}

	data.Username = username
	c := session.DB(b.dbName).C(tableName)
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, &data)
	if b.HandleError(err, w) {
		return
	}

	b.SendJSON(data, w)
}

func (b BuildingsController) delete(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	vars := mux.Vars(r)
	id := vars["id"]

	c := session.DB(b.dbName).C(tableName)
	err := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if b.HandleError(err, w) {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
