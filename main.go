package main

import (
	"log"
	"net/http"
	"time"

	"github.com/akkgr/estiaServer/adapters"
	"github.com/akkgr/estiaServer/auth"
	"github.com/akkgr/estiaServer/controllers"
	"github.com/akkgr/estiaServer/repositories"
	"github.com/gorilla/mux"

	mgo "gopkg.in/mgo.v2"
)

var (
	dbServer   = "mongodb://admin:Adm.123@ds243085.mlab.com:43085/estiag"
	dbName     = "estiag"
	signingKey = []byte("TooSlowTooLate4u.")
	issuer     = "estia"
)

func main() {

	session, err := mgo.Dial(dbServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	db := repositories.NewContext(dbName)
	db.EnsureIndex(session)
	auth.EnsureAdminUser(session, dbName)

	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	h := adapters.Adapt(
		auth.Login(dbName,
			issuer,
			signingKey,
			adapters.DbContextKey),
		adapters.WithDB(session),
		adapters.WithLog(),
		adapters.WithCors())
	authRouter.Handle("/login", h).Methods("POST")

	apiRouter := router.PathPrefix("/api").Subrouter()
	f := new(controllers.FilesController)
	b := new(controllers.BuildingsController)
	routes := append(b.GetRoutes(), f.GetRoutes()...)
	for _, route := range routes {
		h := adapters.Adapt(
			auth.WithAuth(signingKey,
				adapters.UserContextKey,
				route.Handler),
			adapters.WithDB(session),
			adapters.WithCors(),
			adapters.WithLog())
		apiRouter.Handle(route.Path, h).Methods(route.Method)
	}

	staticRouter := router.PathPrefix("/").Subrouter()
	staticRouter.Handle("/{_:.*}", adapters.Adapt(
		adapters.Static("wwwroot", "index.html"),
		adapters.WithCors(),
		adapters.WithLog()))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
