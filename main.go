package main

import (
	"context"
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
