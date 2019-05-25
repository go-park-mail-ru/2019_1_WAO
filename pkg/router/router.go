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
)

//CreateRouter make router consist of 2 part Gorilla Mux and standart router
func CreateRouter(prefix string, serviceSession auth.AuthCheckerClient, db *driver.DB, setting *aws.ConnectSetting) http.Handler {
	userHandler = handlers.NewUserHandler(db, serviceSession, setting)
	Auth = serviceSession
	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix(prefix).Subrouter()

	apiV1.HandleFunc("/users", userHandler.GetAll).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users", userHandler.AddUser).Methods("POST", "OPTIONS")	
	apiV1.Handle("/users/{login}", authMiddleware(http.HandlerFunc(userHandler.GetUsersByNick))).Methods("GET", "OPTIONS")
	apiV1.Handle("/users/{login}", authMiddleware(http.HandlerFunc(userHandler.ModifiedUser))).Methods("PUT", "OPTIONS")
	apiV1.Handle("/session", authMiddleware(http.HandlerFunc(userHandler.Signout))).Methods("DELETE", "OPTIONS")
	apiV1.HandleFunc("/session", userHandler.CheckSession).Methods("GET", "OPTIONS")
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	siteMux := http.NewServeMux()
	siteMux.Handle(prefix + "/", apiV1)
	siteMux.HandleFunc("/signin", userHandler.Signin)
	siteMux.Handle("/favicon.ico", http.NotFoundHandler())

	siteMux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("WAO team"))
	})

	siteMux.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		var signUpForm = []byte(`
		<html>
			<body>
			<form action="/api/users/man/Hotman" method="post" enctype="multipart/form-data">
				NewEmail: <input type="text" name="email">
				NewLogin: <input type="text" name="nickname">
				NewPass: <input type="password" name="password">
				Image: <input type="file" name="image">
				<input type="submit" value="Upd">
			</form>
			</body>
		</html>
		`)
		w.Write(signUpForm)
		return
	})

	siteHandler := CORSMiddleware(siteMux)
	siteHandler = logMiddleware(siteHandler)
	siteHandler = panicMiddleware(siteHandler)
	return siteHandler
}