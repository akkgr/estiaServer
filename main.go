package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akkgr/estiaServer/adapters"
	"github.com/akkgr/estiaServer/controllers"
	"github.com/akkgr/estiaServer/repositories"
	"github.com/gorilla/mux"

	mgo "gopkg.in/mgo.v2"
)

func staticHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	fileSystemPath := "wwwroot" + r.URL.Path
	endURIPath := strings.Split(requestPath, "/")[len(strings.Split(requestPath, "/"))-1]
	splitPath := strings.Split(endURIPath, ".")
	if len(splitPath) > 1 {
		if f, err := os.Stat(fileSystemPath); err == nil && !f.IsDir() {
			http.ServeFile(w, r, fileSystemPath)
			return
		}
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "wwwroot/index.html")
}

func main() {

	session, err := mgo.Dial("mongodb://admin:Adm.123@ds243085.mlab.com:43085/estiag")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	repositories.EnsureIndex(session)
	repositories.EnsureAdminUser(session)

	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	auth := controllers.AuthController{}
	for _, route := range auth.GetRoutes() {
		h := adapters.Adapt(route.Handler, adapters.WithDB(session), adapters.WithLog(), adapters.WithCors())
		authRouter.Handle(route.Path, h).Methods(route.Method)
	}

	apiRouter := router.PathPrefix("/api").Subrouter()
	f := controllers.FilesController{}
	b := controllers.BuildingsController{}
	routes := append(b.GetRoutes(), f.GetRoutes()...)
	for _, route := range routes {
		h := adapters.Adapt(route.Handler, adapters.WithAuth(), adapters.WithDB(session), adapters.WithLog(), adapters.WithCors())
		apiRouter.Handle(route.Path, h).Methods(route.Method)
	}

	router.PathPrefix("/{_:.*}").HandlerFunc(staticHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
