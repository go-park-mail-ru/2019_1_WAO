package router

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/DmitriyPrischep/backend-WAO/pkg/handlers"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/aws"
)

var (
	Auth auth.AuthCheckerClient
	userHandler *handlers.Handler
	frontURL string
)

//CreateRouter make router consist of 2 part Gorilla Mux and standart router
func CreateRouter(prefix, urlCORS, urlImage string, serviceSession auth.AuthCheckerClient, db *driver.DB, setting *aws.ConnectSetting) http.Handler {
	userHandler = handlers.NewUserHandler(db, serviceSession, setting, urlImage)
	Auth = serviceSession
	frontURL = urlCORS
	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix(prefix).Subrouter()

	apiV1.HandleFunc("/users", userHandler.GetAll).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users", userHandler.AddUser).Methods("POST", "OPTIONS")	
	apiV1.Handle("/users/{login}", AuthMiddleware(http.HandlerFunc(userHandler.GetUsersByNick))).Methods("GET", "OPTIONS")
	apiV1.Handle("/users/{login}", AuthMiddleware(http.HandlerFunc(userHandler.ModifiedUser))).Methods("PUT", "OPTIONS")
	apiV1.Handle("/session", AuthMiddleware(http.HandlerFunc(userHandler.Signout))).Methods("DELETE")
	apiV1.HandleFunc("/session", userHandler.CheckSession).Methods("GET", "OPTIONS")

	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	
	siteMux := http.NewServeMux()
	siteMux.Handle(prefix + "/", apiV1)
	siteMux.HandleFunc("/signin", userHandler.Signin)
	siteMux.Handle("/favicon.ico", http.NotFoundHandler())

	siteHandler := CORSMiddleware(siteMux)
	siteHandler = LogMiddleware(siteHandler)
	siteHandler = PanicMiddleware(siteHandler)
	return siteHandler
}