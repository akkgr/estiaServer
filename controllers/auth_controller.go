package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/akkgr/estiaServer/models"
	jwt "github.com/dgrijalva/jwt-go"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// UsersController struct
type AuthController struct {
	Controller
}

// GetRoutes ...
func (u AuthController) GetRoutes() []Route {
	return []Route{
		Route{
			Method:  http.MethodPost,
			Path:    "/login",
			Handler: u.login,
		},
	}
}

func (u AuthController) login(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(dbContextKey).(*mgo.Session)

	var credentials models.User
	err := u.GetContent(&credentials, r)
	if u.HandleError(err, w) {
		return
	}

	var user models.User
	c := session.DB(dbName).C("users")
	err = c.Find(bson.M{"username": strings.ToLower(credentials.Username)}).One(&user)
	if u.HandleError(err, w) {
		return
	}

	if user.Username == "" {
		http.Error(w, "Username not found", http.StatusInternalServerError)
		return
	}

	match := checkPassword(credentials.Password, user.Password)
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
	tokenString, err := token.SignedString(signKey)
	if u.HandleError(err, w) {
		return
	}

	response := jwtToken{Token: tokenString}
	u.SendJSON(response, w)
}
