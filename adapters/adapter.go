package adapters

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/akkgr/estiaServer/repositories"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	mgo "gopkg.in/mgo.v2"
)

// Adapter ...
type Adapter func(http.Handler) http.Handler

// ContextKey ...
type ContextKey string

// DbKey ...
var (
	UserContextKey   = ContextKey("sub")
	DbContextKey     = ContextKey("dbsession")
	IDContextKey     = ContextKey("id")
	OffsetContextKey = ContextKey("offset")
	LimitContextKey  = ContextKey("limit")
)

// Adapt ...
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

// WithLog ...
func WithLog() Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
			h.ServeHTTP(w, r)
		})
	}
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
			ctx := context.WithValue(r.Context(), DbContextKey, dbsession)
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
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					ctx := context.WithValue(r.Context(), UserContextKey, claims["sub"])
					next.ServeHTTP(w, r.WithContext(ctx))
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
