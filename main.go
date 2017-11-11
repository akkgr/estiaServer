package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/akkgr/estiaServer/adapters"
	"github.com/akkgr/estiaServer/controllers"
	"github.com/akkgr/estiaServer/repositories"

	mgo "gopkg.in/mgo.v2"
)

func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type app struct {
	session *mgo.Session
	auth    *auth
	api     *api
}

func (h *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "auth":
		s := adapters.Adapt(h.auth, adapters.WithCors(), adapters.WithDB(h.session))
		s.ServeHTTP(w, r)
	case "api":
		s := adapters.Adapt(h.api, adapters.WithCors(), adapters.WithDB(h.session), adapters.WithAuth())
		s.ServeHTTP(w, r)
	default:
		staticHandler(w, r)
	}
}

type auth struct{}

func (h *auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "login":
		controllers.Login(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

type api struct{}

func (h *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var head string
	head, r.URL.Path = shiftPath(r.URL.Path)
	switch head {
	case "buildings":
		buildingsHandler(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
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

func buildingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		controllers.AllBuildings(w, r)
	case "POST":
		controllers.AddBuild(w, r)
	case "PUT":
		controllers.UpdateBuild(w, r)
	case "DELETE":
		controllers.DeleteBuild(w, r)
	default:
		http.Error(w, "Only GET, POST, PUT and DELETE are allowed", http.StatusMethodNotAllowed)
	}
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

	// router := mux.NewRouter()

	// authRouter := router.PathPrefix("/auth").Subrouter()
	// authRouter.Path("/login").Methods("POST").HandlerFunc(controllers.Login)

	// apiRouter := mux.NewRouter()
	// apiSub := apiRouter.PathPrefix("/api").Subrouter()
	// buildRouter := apiSub.PathPrefix("/buildings").Subrouter()
	// buildRouter.Path("/{offset}/{limit}").Methods("GET").HandlerFunc(controllers.AllBuildings)
	// buildRouter.Path("/{id}").Methods("GET").HandlerFunc(controllers.BuildByID)
	// buildRouter.Path("/").Methods("POST").HandlerFunc(controllers.AddBuild)
	// buildRouter.Path("/{id}").Methods("PUT").HandlerFunc(controllers.UpdateBuild)
	// buildRouter.Path("/{id}").Methods("DELETE").HandlerFunc(controllers.DeleteBuild)
	// h := adapters.Adapt(apiRouter, adapters.WithDB(session), adapters.WithAuth())
	// router.Handle("/api/{_:.*}", h)

	// router.PathPrefix("/{_:.*}").HandlerFunc(staticHandler)

	// headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	// originsOk := handlers.AllowedOrigins([]string{"*"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	// corsRouter := handlers.CORS(originsOk, headersOk, methodsOk)(router)
	// logRouter := handlers.LoggingHandler(os.Stdout, corsRouter)

	app := &app{
		session: session,
		auth:    new(auth),
		api:     new(api),
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      app,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
