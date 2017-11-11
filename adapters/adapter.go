package adapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/akkgr/estiaServer/repositories"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	mgo "gopkg.in/mgo.v2"
)

// Adapter ...
type Adapter func(http.Handler) http.Handler

// DbContextKey ...
type DbContextKey string

// DbKey ...
var DbKey = DbContextKey("dbsession")

// Adapt ...
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

// WithCors ...
func WithCors() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
				w.WriteHeader(http.StatusOK)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}

// WithDB ...
func WithDB(db *mgo.Session) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dbsession := db.Copy()
			defer dbsession.Close()
			ctx := context.WithValue(r.Context(), DbKey, dbsession)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithAuth ...
func WithAuth() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				return repositories.MySigningKey, nil
			})

			if err == nil {
				if token.Valid {
					next.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprint(w, "Token is not valid")
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Unauthorised access to this resource")
			}
		})
	}
}
