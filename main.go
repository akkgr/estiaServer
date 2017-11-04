package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"

	mgo "gopkg.in/mgo.v2"
)

var db = "estia"

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

func corsMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		w.WriteHeader(http.StatusOK)
	} else {
		next(w, r)
	}
}

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
		log.Print(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)
	ensureAdminUser(session)

	router := mux.NewRouter()

	authBase := mux.NewRouter()
	router.PathPrefix("/auth").Handler(negroni.New(
		negroni.HandlerFunc(corsMiddleware),
		negroni.Wrap(authBase),
	))
	authRouter := authBase.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST").HandlerFunc(login(session))

	apiBase := mux.NewRouter()
	router.PathPrefix("/api").Handler(negroni.New(
		negroni.HandlerFunc(corsMiddleware),
		negroni.HandlerFunc(authMiddleware),
		negroni.Wrap(apiBase),
	))
	apiRouter := apiBase.PathPrefix("/api").Subrouter()
	apiRouter.Path("/buildings/{offset}/{limit}").Methods("GET").HandlerFunc(allBuildings(session))
	apiRouter.Path("/buildings/{id}").Methods("GET").HandlerFunc(buildByID(session))
	apiRouter.Path("/buildings").Methods("POST").HandlerFunc(addBuild(session))
	apiRouter.Path("/buildings/{id}").Methods("PUT").HandlerFunc(updateBuild(session))
	apiRouter.Path("/buildings/{id}").Methods("DELETE").HandlerFunc(deleteBuild(session))

	router.PathPrefix("/{_:.*}").HandlerFunc(staticHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
