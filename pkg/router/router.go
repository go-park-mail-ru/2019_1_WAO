package router

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/DmitriyPrischep/backend-WAO/pkg/handlers"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
)

var (
	PathStaticServer string
	Auth auth.AuthCheckerClient
)

func CreateRouter(prefix, pathToStaticFiles string, serviceSession auth.AuthCheckerClient, db *driver.DB) *http.ServeMux {
	userHandler := handlers.NewUserHandler(db, serviceSession)
	PathStaticServer = pathToStaticFiles
	Auth = serviceSession
	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix(prefix).Subrouter()

	apiV1.HandleFunc("/users", userHandler.GetAll).Methods("GET", " OPTIONS")
	apiV1.HandleFunc("/users", userHandler.AddUser).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", userHandler.GetUsersByNick).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", userHandler.ModifiedUser).Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/session", userHandler.Signout).Methods("DELETE", "OPTIONS")
	apiV1.HandleFunc("/session", userHandler.CheckSession).Methods("GET", "OPTIONS")
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	siteMux := http.NewServeMux()
	siteMux.Handle(prefix + "/", apiV1)
	siteMux.HandleFunc("/signin", userHandler.Signin)
	siteMux.Handle("/favicon.ico", http.NotFoundHandler())

	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir(pathToStaticFiles)),
	)
	siteMux.Handle("/data/", staticHandler)

	return siteMux
}