package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type jwtToken struct {
	Token string `json:"token"`
}

// Login ...
func Login(dbName string,
	issuer string,
	key []byte,
	dbContextKey interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value(dbContextKey).(*mgo.Session)

		var credentials User
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&credentials)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var user User
		c := session.DB(dbName).C("users")
		err = c.Find(bson.M{"username": strings.ToLower(credentials.Username)}).One(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user.Username == "" {
			http.Error(w, "Username not found", http.StatusInternalServerError)
			return
		}

		match := CheckPasswordHash(credentials.Password, user.Password)
		if match == false {
			http.Error(w, "Invalid credentials", http.StatusInternalServerError)
			return
		}

		exp := time.Now().Add(time.Hour * 8).Unix()
		claims := &jwt.StandardClaims{
			ExpiresAt: exp,
			Issuer:    issuer,
			Subject:   user.Username,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := jwtToken{Token: tokenString}
		json, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	})
}
