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
	"github.com/gorilla/handlers"
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
	authRouter.Path("/login").Methods("POST").HandlerFunc(controllers.Login)

	apiRouter := mux.NewRouter()
	apiSub := apiRouter.PathPrefix("/api").Subrouter()
	buildRouter := apiSub.PathPrefix("/buildings").Subrouter()
	buildRouter.Path("/{offset}/{limit}").Methods("GET").HandlerFunc(controllers.AllBuildings)
	buildRouter.Path("/{id}").Methods("GET").HandlerFunc(controllers.BuildByID)
	buildRouter.Path("/").Methods("POST").HandlerFunc(controllers.AddBuild)
	buildRouter.Path("/{id}").Methods("PUT").HandlerFunc(controllers.UpdateBuild)
	buildRouter.Path("/{id}").Methods("DELETE").HandlerFunc(controllers.DeleteBuild)
	h := adapters.Adapt(apiRouter, adapters.WithDB(session), adapters.WithAuth())
	router.Handle("/api/{_:.*}", h)

	router.PathPrefix("/{_:.*}").HandlerFunc(staticHandler)

	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	corsRouter := handlers.CORS(originsOk, headersOk, methodsOk)(router)
	logRouter := handlers.LoggingHandler(os.Stdout, corsRouter)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      logRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
