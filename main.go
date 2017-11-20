package main

import (
	"context"
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

type app struct {
	session *mgo.Session
	router  *mux.Router
	auth    *auth
	api     *api
}

func (h *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := shiftPath(r.URL.Path)
	switch head {
	case "auth":
		r.URL.Path = tail
		s := adapters.Adapt(h.auth, adapters.WithDB(h.session), adapters.WithLog(), adapters.WithCors())
		s.ServeHTTP(w, r)
	case "api":
		r.URL.Path = tail
		s := adapters.Adapt(h.api, adapters.WithAuth(), adapters.WithDB(h.session), adapters.WithLog(), adapters.WithCors())
		s.ServeHTTP(w, r)
	default:
		staticHandler(w, r)
	}

	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.Path("/login").Methods("POST").HandlerFunc(controllers.Login)

	apiRouter := mux.NewRouter()
	apiSub := apiRouter.PathPrefix("/api").Subrouter()
	buildRouter := apiSub.PathPrefix("/buildings").Subrouter()
	for _, route := range b.GetRoutes() {
		s.HandleFunc(route.Path, route.Handler).Methods(route.Method)
	}
	h := adapters.Adapt(apiRouter, adapters.WithDB(session), adapters.WithAuth())
	router.Handle("/api/{_:.*}", h)

	router.PathPrefix("/{_:.*}").HandlerFunc(staticHandler)
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
	case "files":
		fileHandler(w, r)
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

func fileHandler(w http.ResponseWriter, r *http.Request) {
	head, tail := shiftPath(r.URL.Path)
	switch r.Method {
	case "GET":
		if tail == "/" {
			ctx := context.WithValue(r.Context(), adapters.IDContextKey, head)
			controllers.DownloadFile(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Only GET, POST and DELETE are allowed", http.StatusMethodNotAllowed)
		}
	case "POST":
		controllers.UploadFile(w, r)
	case "DELETE":
		http.Error(w, "Only GET, POST and DELETE are allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Only GET, POST and DELETE are allowed", http.StatusMethodNotAllowed)
	}
}

func buildingsHandler(w http.ResponseWriter, r *http.Request) {
	head, tail := shiftPath(r.URL.Path)
	switch r.Method {
	case "GET":
		if tail == "/" {
			ctx := context.WithValue(r.Context(), adapters.IDContextKey, head)
			controllers.BuildByID(w, r.WithContext(ctx))
		} else {
			ctx := context.WithValue(r.Context(), adapters.OffsetContextKey, head)
			ctx = context.WithValue(ctx, adapters.LimitContextKey, tail)
			controllers.AllBuildings(w, r.WithContext(ctx))
		}
	case "POST":
		controllers.AddBuild(w, r)
	case "PUT":
		ctx := context.WithValue(r.Context(), adapters.IDContextKey, head)
		controllers.UpdateBuild(w, r.WithContext(ctx))
	case "DELETE":
		ctx := context.WithValue(r.Context(), adapters.IDContextKey, head)
		controllers.DeleteBuild(w, r.WithContext(ctx))
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

	app := &app{
		session: session,
		router:  mux.NewRouter().StrictSlash(true),
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
