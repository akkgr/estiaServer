package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

var mySigningKey = []byte("TooSlowTooLate4u.")

type jwtToken struct {
	Token string `json:"token"`
}

type userCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var credentials userCredentials
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var user user
		c := session.DB(db).C("users")
		err = c.Find(bson.M{"username": strings.ToLower(credentials.Username)}).One(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if user.Username == "" {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		if user.Password != credentials.Password {
			http.Error(w, "Invalid credentials", http.StatusInternalServerError)
			return
		}

		exp := time.Now().Add(time.Minute * 20).Unix()
		claims := &jwt.StandardClaims{
			ExpiresAt: exp,
			Issuer:    "estia",
			Subject:   user.Username,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(mySigningKey)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error while signing the token")
			log.Printf("Error signing token: %v\n", err)
		}

		response := jwtToken{tokenString}
		jsonResponse(response, w)
	}
}

func authMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err == nil {
		if token.Valid {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorised access to this resource")
	}
}
