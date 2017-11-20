package adapters

import (
	"context"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

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
