package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/DmitriyPrischep/backend-WAO/docs"
	_ "github.com/lib/pq"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type User struct {
	ID       int    `json:"id, string, omitempty"`
	Email    string `json:"email, omitempty"`
	password string `json:"password, omitempty"`
	Nick     string `json:"nick, omitempty"`
	Scope    int    `json:"scope, string, omitempty"`
	Games    int    `json:"games, string, omitempty"`
	Wins     int    `json:"wins, string, omitempty"`
	Image    string `json:"image, omitempty"`
}

type Error struct {
	Message string
}

var secret = []byte("secretkey")
var db *sql.DB

func checkAuthorization(r http.Request) (jwt.MapClaims, bool, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, false, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, err
		}
		return secret, nil
	})
	if err != nil {
		log.Printf("Unexpected signing method: %v", token.Header["alg"])
		return nil, false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true, nil
	}
	return nil, false, nil
}

// ShowAccount godoc
// @Summary Show statistic of users
// @Description get all users
// @Accept  json
// @Produce  json
// @Header 200 {object} User
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /users [get]
func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, email, nickname, scope, games, wins, image FROM users")
	if err != nil {
		log.Println("Method GetUsers:", err)
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Nick, &user.Scope, &user.Games, &user.Wins, &user.Image)
		if err != nil {
			log.Println(err)
			continue
		}
		if user.Image != "" {
			user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
		}
		users = append(users, user)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ShowAccount godoc
// @Summary Show data of user
// @Description get user by ID
// @ID get-user-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /api/users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	row := db.QueryRow(`SELECT id, email, nickname, scope, games, wins, image 
						FROM users WHERE nickname = $1`, params["login"])

	user := User{}

	switch err := row.Scan(&user.ID, &user.Email, &user.Nick, &user.Scope,
		&user.Games, &user.Wins, &user.Image); err {
	case sql.ErrNoRows:
		log.Println("Method GetUser: No rows were returned!")
	case nil:
		if user.Image != "" {
			user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
		}
		json.NewEncoder(w).Encode(user)
		return
	default:
		log.Println("Method GetUser: ", err)
	}
	http.Error(w, `{"error": "This user is not found}"`, http.StatusNotFound)
}

// ShowAccount godoc
// @Summary Add user
// @Description Add user in DB
// @Accept  json
// @Produce  json
// @Param default query string false "string default" default(A)
// @Success 200 {object} User
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /api/users [post]
func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Decode error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Email == "" || user.password == "" {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"error": "Invalid data"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id := 0
	err = db.QueryRow(`INSERT INTO users (email, nickname, password, scope, games, wins, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		user.Email, user.Nick, user.password, 0, 0, 0, "").Scan(&id)
	if err != nil {
		log.Printf("Error inserting record: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("New record ID is:", id)
	user.ID = id
	json.NewEncoder(w).Encode(user)

}

func uploadAvatar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		log.Println(err.Error())
	}
	file, handler, err := r.FormFile("input_file")
	if err != nil {
		fmt.Println(err)
	} else {
		defer file.Close()
		if _, err := os.Stat("./static"); os.IsNotExist(err) {
			err = os.Mkdir("./static", 0700)
			if err != nil {
				log.Println(err.Error())
			}
		}

		user := User{}
		dirname := strconv.Itoa(user.ID)
		if _, err := os.Stat("./static/" + dirname); os.IsNotExist(err) {
			err = os.Mkdir("./static/"+dirname, 0400)
			if err != nil {
				log.Println(err.Error())
			}
		}
		if err != nil {
			log.Println(err.Error())
		}

		saveFile, err := os.Create("./static/" + dirname + "/" + handler.Filename)
		if err != nil {
			log.Println(err.Error())
		}
		defer saveFile.Close()

		_, err = io.Copy(saveFile, file)
		if err != nil {
			log.Println(err.Error())
		}
		user.Image = handler.Filename
	}
}

// ShowAccount godoc
// @Summary Update user data
// @Description Update user data by ID
// @ID get-user-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /api/users/{id} [put]
func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	userLogin := params["login"]
	if userLogin == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	user := User{}
	err := db.QueryRow(`UPDATE users
		SET email = $1, nickname = $2
		WHERE nickname = $2 RETURNING id, email, nickname, scope, games, wins, image;`,
		"new_email", "new_nick").Scan(user.ID, user.Email, user.Nick, user.Scope, user.Games, user.Wins, user.Image)
	if err != nil {
		log.Println("Error updating record:", err)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	json.NewEncoder(w).Encode(user)
	w.WriteHeader(http.StatusOK)
}

// ShowAccount godoc
// @Summary Delete data of user
// @Description remove user by ID
// @ID get-user-by-int
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Header 200 {string} Location "/"
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /users/{id} [delete]
func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	_, err := db.Exec(`DELETE FROM users
	WHERE id = $1;`, params["login"])
	if err != nil {
		log.Println("Error deleting record:", err)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func signin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signin" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	row := db.QueryRow(`SELECT email, nickname, password 
						FROM users WHERE nickname = $1 AND password = $2`, r.FormValue("login"), r.FormValue("password"))

	user := User{}

	switch err := row.Scan(&user.Email, &user.Nick, &user.password); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		http.Error(w, `{"error": "Invalid login or password"}`, http.StatusBadRequest)
		return
	case nil:
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": r.FormValue("login"),
			"exp":      time.Now().Add(1 * time.Minute).Unix(),
		})

		token, err := rawToken.SignedString(secret)
		if err != nil {
			w.Write([]byte("Error: Token was not create!" + err.Error()))
			return
		}

		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    token,
			Expires:  time.Now().Add(10 * time.Minute),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		w.Write([]byte("You are authorized! Welcome!"))
		return
	default:
		log.Println("Method GetUser: ", err)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func redirectOnMain(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/", http.StatusSeeOther)
	return
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WAO team"))
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	if r.Method != http.MethodGet {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}
	_, err := r.Cookie("session_id")
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	expiredCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().AddDate(0, -1, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, expiredCookie)
	w.WriteHeader(http.StatusOK)
}
func errorHandler(w http.ResponseWriter, r *http.Request, code int) {
	if code == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	http.Error(w, "", code)
}

func checkSession(w http.ResponseWriter, r *http.Request) {
	claims, ok, err := checkAuthorization(*r)
	if !ok {
		log.Println("Autorization checking error:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, ok = claims["username"]
	if !ok {
		log.Println("Bad claims: field 'username' not exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func authMiddleware(next http.Handler) http.Handler {
	log.Println("Something")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok, err := checkAuthorization(*r); !ok {
			log.Println(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			// http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		log.Println("Complete check auth")
		next.ServeHTTP(w, r)
		return
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("logMiddleware", r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s, %s %s\n",
			r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovered", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func init() {
	connStr := "user=postgres password=123456 dbname=waogame sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
	}
}

// @title Web Art Online API
// @version 1.0
// @description This is a game server.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /api/
func main() {
	defer db.Close()

	actionMux := mux.NewRouter()

	apiV1 := actionMux.PathPrefix("/api").Subrouter()
	apiV1.HandleFunc("/users", GetUsers).Methods("GET")
	apiV1.HandleFunc("/users/{login}", GetUser).Methods("GET")
	apiV1.HandleFunc("/users", CreateUser).Methods("POST")
	apiV1.HandleFunc("/users/{login}", updateUser).Methods("PUT")
	apiV1.HandleFunc("/users/{login}", deleteUser).Methods("DELETE")
	apiV1.HandleFunc("/session", checkSession).Methods("GET")
	apiV1.HandleFunc("/session", logout).Methods("DELETE")
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Ппссс, парень! Такой страницы не существует!"))
	})

	apiV1.Use(authMiddleware)

	siteMux := http.NewServeMux()
	siteMux.Handle("/api/", apiV1)
	siteMux.HandleFunc("/docs/", httpSwagger.WrapHandler)
	siteMux.HandleFunc("/", mainPage)
	siteMux.HandleFunc("/signin", signin)
	siteMux.HandleFunc("/logout", logout)

	siteMux.HandleFunc("/gets", GetUsers)
	siteMux.HandleFunc("/get", GetUser)
	siteMux.HandleFunc("/add", CreateUser)
	siteMux.HandleFunc("/del", deleteUser)

	siteMux.Handle("/favicon.ico", http.NotFoundHandler())

	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir("./static")),
	)
	siteMux.Handle("/data/", staticHandler)

	siteHandler := logMiddleware(siteMux)
	siteHandler = panicMiddleware(siteHandler)

	srv := &http.Server{
		Handler:      siteHandler,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server at http://127.0.0.1:8000/")
	log.Println(srv.ListenAndServe())
}
