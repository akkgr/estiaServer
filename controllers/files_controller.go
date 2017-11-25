package controllers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FilesController struct
type FilesController struct {
	Controller
}

// GetRoutes ...
func (c FilesController) GetRoutes() []Route {
	return []Route{
		Route{
			Method:  http.MethodGet,
			Path:    "/files/{id}",
			Handler: c.download,
		},
		Route{
			Method:  http.MethodPost,
			Path:    "/files",
			Handler: c.upload,
		},
	}
}

func (c FilesController) upload(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)
	gfs := session.DB(dbName).GridFS("fs")

	file, handler, err := r.FormFile("file")
	if c.HandleError(err, w) {
		return
	}
	defer file.Close()

	limit := int64(30 * 1024 * 1024)
	if handler.Size > limit {
		http.Error(w, "Maximum File size is 30Mb", http.StatusInternalServerError)
		return
	}

	h := md5.New()
	_, err = io.Copy(h, file)
	if c.HandleError(err, w) {
		return
	}

	md5hash := fmt.Sprintf("%x", h.Sum(nil))
	var old []*mgo.GridFile
	err = gfs.Find(bson.M{"md5": md5hash}).All(&old)
	if c.HandleError(err, w) {
		return
	}

	if len(old) > 0 {
		c.SendJSON(old[0].Id(), w)
		return
	}

	dbfile, err := gfs.Create(handler.Filename)
	if c.HandleError(err, w) {
		return
	}
	_, err = io.Copy(dbfile, file)
	if c.HandleError(err, w) {
		return
	}

	err = dbfile.Close()
	if c.HandleError(err, w) {
		return
	}

	c.SendJSON(dbfile.Id(), w)
}

func (c FilesController) download(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	vars := mux.Vars(r)
	id := vars["id"]

	dbfile, err := session.DB(dbName).GridFS("fs").OpenId(bson.ObjectIdHex(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(w, dbfile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dbfile.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+dbfile.Name())
	w.Header().Set("Content-Type", "application/x-download")
}
