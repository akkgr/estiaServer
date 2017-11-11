package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/akkgr/estiaServer/controllers"
	"github.com/akkgr/estiaServer/repositories"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	mgo "gopkg.in/mgo.v2"
)

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return repositories.MySigningKey, nil
		})

		if err == nil {
			if token.Valid {
				next.ServeHTTP(w, r)
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
	authRouter.Path("/login").Methods("POST").HandlerFunc(controllers.Login(session))

	apiRouter := mux.NewRouter()
	apiSub := apiRouter.PathPrefix("/api").Subrouter()
	buildRouter := apiSub.PathPrefix("/buildings").Subrouter()
	buildRouter.Path("/{offset}/{limit}").Methods("GET").HandlerFunc(controllers.AllBuildings(session))
	buildRouter.Path("/{id}").Methods("GET").HandlerFunc(controllers.BuildByID(session))
	buildRouter.Path("/").Methods("POST").HandlerFunc(controllers.AddBuild(session))
	buildRouter.Path("/{id}").Methods("PUT").HandlerFunc(controllers.UpdateBuild(session))
	buildRouter.Path("/{id}").Methods("DELETE").HandlerFunc(controllers.DeleteBuild(session))
	router.Handle("/api/{_:.*}", auth(apiRouter))

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
