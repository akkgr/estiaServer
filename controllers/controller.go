package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akkgr/estiaServer/adapters"
	"github.com/akkgr/estiaServer/models"
	"github.com/akkgr/estiaServer/repositories"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db = repositories.Db

func jsonResponse(response interface{}, w http.ResponseWriter) {
	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

// Login ...
func Login(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	var credentials models.User
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user models.User
	c := session.DB(db).C("users")
	err = c.Find(bson.M{"username": strings.ToLower(credentials.Username)}).One(&user)
	if err != nil {
		http.Error(w, "Username not found", http.StatusInternalServerError)
		return
	}

	if user.Username == "" {
		http.Error(w, "Username not found", http.StatusInternalServerError)
		return
	}

	match := repositories.CheckPasswordHash(credentials.Password, user.Password)
	if match == false {
		http.Error(w, "Invalid credentials", http.StatusInternalServerError)
		return
	}

	exp := time.Now().Add(time.Hour * 8).Unix()
	claims := &jwt.StandardClaims{
		ExpiresAt: exp,
		Issuer:    "estia",
		Subject:   user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(repositories.MySigningKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error while signing the token")
		log.Printf("Error signing token: %v\n", err)
	}

	response := models.JwtToken{Token: tokenString}
	jsonResponse(response, w)
}

// AllBuildings ...
func AllBuildings(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	vars := mux.Vars(r)
	offset, _ := strconv.Atoi(vars["offset"])
	limit, _ := strconv.Atoi(vars["limit"])

	c := session.DB(db).C("buildings")

	var data []models.Building
	count, err := c.Find(bson.M{}).Count()
	err = c.Find(bson.M{}).Sort(
		"address.street",
		"address.streetNumber",
		"address.area",
		"address.country").Skip(offset).Limit(limit).All(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp models.DataResponse
	resp.Data = data
	resp.Count = count
	jsonResponse(resp, w)
}

// BuildByID ...
func BuildByID(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	var data models.Building

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "0" {
		jsonResponse(data, w)
		return
	}

	c := session.DB(db).C("buildings")
	err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(data, w)
}

// AddBuild ...
func AddBuild(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	data := models.Building{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := session.DB(db).C("buildings")

	err := c.Insert(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(data, w)
}

// UpdateBuild ...
func UpdateBuild(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	vars := mux.Vars(r)
	id := vars["id"]

	data := models.Building{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := session.DB(db).C("buildings")

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(id)}, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(data, w)
}

// DeleteBuild ...
func DeleteBuild(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(adapters.DbKey).(*mgo.Session)

	vars := mux.Vars(r)
	id := vars["id"]

	c := session.DB(db).C("buildings")

	err := c.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}