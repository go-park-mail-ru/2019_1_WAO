package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type UserRegister struct {
	Email    string `json:"email, omitempty"`
	Password string `json:"password, omitempty"`
	Nickname string `json:"nickname, omitempty"`
}

type User struct {
	ID       int    `json:"id, string, omitempty"`
	Email    string `json:"email, omitempty"`
	password string `json:"password, omitempty"`
	Nick     string `json:"nickname, omitempty"`
	Score    int    `json:"score, string, omitempty"`
	Games    int    `json:"games, string, omitempty"`
	Wins     int    `json:"wins, string, omitempty"`
	Image    string `json:"image, omitempty"`
}

type signinUser struct {
	Nickname string `json:"nickname, omitempty"`
	Password string `json:"password, omitempty"`
}

type userNickname struct {
	Nickname string `json:"nickname, omitempty"`
}

type Error struct {
	Message string
}

var db *sql.DB

var (
	sessionManager auth.AuthCheckerClient
)

func GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	rows, err := db.Query("SELECT id, email, nickname, scope, games, wins, image FROM users ORDER by scope DESC;")
	if err != nil {
		log.Println("Method GetUsers:", err)
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Nick, &user.Score, &user.Games, &user.Wins, &user.Image)
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
	w.Write([]byte(`{"users":`))
	json.NewEncoder(w).Encode(users)
	w.Write([]byte(`}`))
}

// func GetUser(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodOptions {
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	params := mux.Vars(r)

// 	row := db.QueryRow(`SELECT id, email, nickname, scope, games, wins, image
// 						FROM users WHERE nickname = $1;`, params["login"])

// 	user := User{}

// 	switch err := row.Scan(&user.ID, &user.Email, &user.Nick, &user.Score,
// 		&user.Games, &user.Wins, &user.Image); err {
// 	case sql.ErrNoRows:
// 		log.Println("Method GetUser: No rows were returned!")
// 	case nil:
// 		if user.Image != "" {
// 			user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
// 		}
// 		json.NewEncoder(w).Encode(user)
// 		return
// 	default:
// 		log.Println("Method GetUser: ", err)
// 	}
// 	http.Error(w, `{"error": "This user is not found"}`, http.StatusNotFound)
// }

// func CreateUser(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == http.MethodOptions {
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	var user UserRegister
// 	err := json.NewDecoder(r.Body).Decode(&user)
// 	if err != nil {
// 		log.Printf("Decode error: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	if user.Email == "" || user.Nickname == "" || user.Password == "" {
// 		w.Header().Set("Content-Type", "application/json")
// 		io.WriteString(w, `{"error": "Uncorrect email or nickname or password"}`)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	var nickname string
// 	err = db.QueryRow(`INSERT INTO users (email, nickname, password, scope, games, wins, image)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING nickname`,
// 		user.Email, user.Nickname, user.Password, 0, 0, 0, "").Scan(&nickname)
// 	if err != nil {
// 		log.Printf("Error inserting record: %v", err)
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	fmt.Println("New record NICK is:", nickname)

// 	sess, err := sessionManager.Create(
// 		context.Background(),
// 		&auth.UserData{
// 			Login: nickname,
// 			Agent: r.UserAgent(),
// 		})
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	cookie := &http.Cookie{
// 		Name:     "session_id",
// 		Value:    sess.Value,
// 		Expires:  time.Now().Add(10 * time.Minute),
// 		HttpOnly: true,
// 	}
// 	http.SetCookie(w, cookie)
// }

func getSession(r *http.Request) (*auth.UserData, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sess, err := sessionManager.Check(
		context.Background(),
		&auth.Token{
			Value: cookieSessionID.Value,
		})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func checkSession(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/session" {
		log.Println(r.URL.Path, "ERROR")
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	if r.Method == http.MethodOptions {
		return
	}
	if r.Method != http.MethodGet {
		errorHandler(w, r, http.StatusNotFound)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Println("Error checking of session")
	}
	if session == nil {
		errorHandler(w, r, http.StatusBadRequest)
		return
	}

	nickname := userNickname{
		Nickname: session.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nickname)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	row := db.QueryRow(`SELECT id, email, nickname, scope, games, wins, image 
						FROM users WHERE nickname = $1;`, params["login"])

	user := User{}

	switch err := row.Scan(&user.ID, &user.Email, &user.Nick, &user.Score,
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
	http.Error(w, `{"error": "This user is not found"}`, http.StatusNotFound)
}

func errorHandler(w http.ResponseWriter, r *http.Request, code int) {
	if code == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	http.Error(w, "", code)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Println("Error checking of session")
	}
	w.Write([]byte("WAO team"))

	if session != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "<br>Welcome "+session.Login)
	}
}

func signout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	val, err := r.Cookie("session_id")
	if err != nil {
		errorHandler(w, r, http.StatusBadRequest)
		log.Println("Error: ", val)
		return
	}
	log.Println("Session: ", val)

	expiredCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().AddDate(0, -1, 0),
		HttpOnly: true,
	}
	http.SetCookie(w, expiredCookie)
	w.WriteHeader(http.StatusOK)
}

func signin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signin" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	// body, err := ioutil.ReadAll(r.Body)
	// defer r.Body.Close()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println("Body: ", string(body))
	// data := signinUser{}
	// if err := json.Unmarshal(body, &data); err != nil {
	// 	log.Println(err)
	// }
	// log.Println("Structure: ", data)

	data := signinUser{
		Nickname: r.FormValue("login"),
		Password: r.FormValue("password"),
	}
	log.Println("User -- ", data)
	token, err := sessionManager.Create(
		context.Background(),
		&auth.UserData{
			Login:    data.Nickname,
			Password: data.Password,
			Agent:    r.UserAgent(),
		})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    token.Value,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	w.Write([]byte("You are authorized! Welcome!"))
	return
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	inputLogin := r.FormValue("login")
	expiration := time.Now().Add(10 * time.Minute)

	sess, err := sessionManager.Create(
		context.Background(),
		&auth.UserData{
			Login: inputLogin,
			Agent: r.UserAgent(),
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sess.Value,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func checkAuthorization(r http.Request) (jwt.MapClaims, bool, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, false, err
	}

	secret := []byte("secretkey")
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

// func init() {
// 	connectStr := "user=postgres password=123456 dbname=waogame sslmode=disable"

// 	userDB := viper.GetString("db.user")
// 	userPass := viper.GetString("db.password")
// 	nameDB := viper.GetString("db.name")
// 	ssl := viper.GetString("db.sslmode")

// 	fmt.Println("INFO: ", userDB, userPass, nameDB, ssl)

// 	connectStr2 := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
// 		viper.GetString("db.user"), viper.GetString("db.password"), viper.GetString("db.name"), viper.GetString("db.sslmode"))

// 	fmt.Printf("Equal:\n%s\n%s\n", connectStr, connectStr2)
// 	var err error
// 	db, err = sql.Open("postgres", connectStr)
// 	if err != nil {
// 		log.Printf("No connection to DB: %v", err)
// 		return
// 	}
// }

func main() {
	viper.AddConfigPath("../../")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}

	userDB := viper.GetString("db.user")
	userPass := viper.GetString("db.password")
	nameDB := viper.GetString("db.name")
	sslMode := viper.GetString("db.sslmode")
	port := viper.GetString("apisrv.port")
	host := viper.GetString("apisrv.host")
	hostAuth := viper.GetString("authsrv.host") + ":" + viper.GetString("authsrv.port")
	fmt.Println("INFO: ", userDB, userPass, nameDB, sslMode)

	connectStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
		userDB, userPass, nameDB, sslMode)
	fmt.Println("Str: ", connectStr)

	var err error
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
		return
	}
	defer db.Close()

	grpcConnect, err := grpc.Dial(
		hostAuth,
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Println("Can't connect to gRPC")
	}
	defer grpcConnect.Close()

	sessionManager = auth.NewAuthCheckerClient(grpcConnect)

	defer db.Close()

	actionMux := mux.NewRouter()
	apiV1 := actionMux.PathPrefix("/api").Subrouter()

	apiV1.HandleFunc("/users", GetUsers).Methods("GET", " OPTIONS")
	// apiV1.HandleFunc("/users", CreateUser).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", GetUser).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/session", signout).Methods("DELETE", "OPTIONS")
	apiV1.HandleFunc("/session", checkSession).Methods("GET", "OPTIONS")
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	siteMux := http.NewServeMux()
	siteMux.Handle("/api/", apiV1)
	siteMux.HandleFunc("/", mainPage)
	siteMux.HandleFunc("/signin", signin)
	siteMux.HandleFunc("/signout", signout)
	http.HandleFunc("/login", loginPage)
	siteMux.Handle("/favicon.ico", http.NotFoundHandler())
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

	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir("./static")),
	)
	siteMux.Handle("/data/", staticHandler)

	srv := &http.Server{
		Handler:      siteMux,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server at http://" + srv.Addr)
	log.Println(srv.ListenAndServe())
}
