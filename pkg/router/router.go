package router

import (
	"log"
	"time"
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

const (
	frontAddres = "http://127.0.0.1:3000"
)

// func authMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("authMiddleware", r.URL.Path)
// 		cookie, _ := r.Cookie("session_id")
// 		log.Println("Token:", cookie)

// 		if _, ok, err := checkAuthorization(*r); !ok {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			log.Println(err.Error())
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("logMiddleware", r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s, %s %s\n",
			r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovered", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("CORSMiddleware", r.URL.Path)
		//HARD URL
		if origin := r.Header.Get("Origin"); origin == frontAddres {
			w.Header().Set("Access-Control-Allow-Origin", frontAddres)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		next.ServeHTTP(w, r)
	})
}

//CreateRouter make router consist of 2 part Gorilla Mux and standart router
func CreateRouter(prefix, pathToStaticFiles string, serviceSession auth.AuthCheckerClient, db *driver.DB, setting *aws.ConnectSetting) http.Handler {
	userHandler := handlers.NewUserHandler(db, serviceSession, setting)
	PathStaticServer = pathToStaticFiles
	Auth = serviceSession
	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix(prefix).Subrouter()

	apiV1.HandleFunc("/users", userHandler.GetAll).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users", userHandler.AddUser).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", userHandler.GetUsersByNick).Methods("GET", "OPTIONS")
	// apiV1.Handle("/users/{login}", authMiddleware(http.HandlerFunc(GetUser))).Methods("GET")
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

	siteHandler := CORSMiddleware(siteMux)
	siteHandler = logMiddleware(siteHandler)
	siteHandler = panicMiddleware(siteHandler)

	return siteHandler
}