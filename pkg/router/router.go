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
	PathStaticServer string
	Auth auth.AuthCheckerClient
)

//CreateRouter make router consist of 2 part Gorilla Mux and standart router
func CreateRouter(prefix, pathToStaticFiles string, serviceSession auth.AuthCheckerClient, db *driver.DB, setting *aws.ConnectSetting) *http.ServeMux {
	userHandler := handlers.NewUserHandler(db, serviceSession, setting)
	PathStaticServer = pathToStaticFiles
	Auth = serviceSession
	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix(prefix).Subrouter()

	apiV1.HandleFunc("/users", userHandler.GetAll).Methods("GET", " OPTIONS")
	apiV1.HandleFunc("/users", userHandler.AddUser).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", userHandler.GetUsersByNick).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users/man/{login}", userHandler.ModifiedUser) //.Methods("PUT", "OPTIONS")
	apiV1.HandleFunc("/session", userHandler.Signout).Methods("DELETE", "OPTIONS")
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
		// session, err := getSession(r)
		// if err != nil {
		// 	log.Println("Error checking of session")
		// }
	
		// if session != nil {
		// 	w.Header().Set("Content-Type", "text/html")
		// 	fmt.Fprintln(w, "\nWelcome "+session.Login)
		// }
		
	})

	siteMux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var loginFormTmpl = []byte(`
		<html>
			<body>
			<form action="/signin" method="post">
				Login: <input type="text" name="login">
				Password: <input type="password" name="password">
				<input type="submit" value="Login">
			</form>
			</body>
		</html>
		`)
		w.Write(loginFormTmpl)
	})
	siteMux.HandleFunc("/reg", func(w http.ResponseWriter, r *http.Request) {
		var signUpForm = []byte(`
		<html>
			<body>
			<form action="/api/users" method="post">
				Email: <input type="text" name="email">
				Login: <input type="text" name="login">
				Password: <input type="password" name="password">
				<input type="submit" value="Reg">
			</form>
			</body>
		</html>
		`)
		w.Write(signUpForm)
		return
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


	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir(pathToStaticFiles)),
	)
	siteMux.Handle("/data/", staticHandler)

	return siteMux
}