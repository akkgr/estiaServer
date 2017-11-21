package adapters

import (
	"net/http"
)

// Adapter ...
type Adapter func(http.Handler) http.Handler

// ContextKey ...
type ContextKey string

// DbKey ...
var (
	UserContextKey = ContextKey("sub")
	DbContextKey   = ContextKey("dbsession")
)

// Adapt ...
func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}
