package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
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

type updateDataImport struct {
	Nickname string `json:"nickname, omitempty"`
	Password string `json:"password, omitempty"`
	Email    string `json:"email, omitempty"`
}

type updateDataExport struct {
	Email    string `json:"email, omitempty"`
	Nickname string `json:"nickname, omitempty"`
	Image    string `json:"image, omitempty"`
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

const (
	expiration        = 10 * time.Minute
	pathToStaticFiles = "./static"
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

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// REPLACE
	// var user UserRegister
	// err := json.NewDecoder(r.Body).Decode(&user)
	// if err != nil {
	// 	log.Printf("Decode error: %v", err)
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	user := UserRegister{
		Email:    r.FormValue("email"),
		Nickname: r.FormValue("login"),
		Password: r.FormValue("password"),
	}

	if user.Email == "" || user.Nickname == "" || user.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"error": "Uncorrect email or nickname or password"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var nickname string
	var id string
	err := db.QueryRow(`INSERT INTO users (email, nickname, password, scope, games, wins, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, nickname`,
		user.Email, user.Nickname, user.Password, 0, 0, 0, "").Scan(&id, &nickname)
	if err != nil {
		log.Printf("Error inserting record: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println("New record NICK is:", nickname)

	sess, err := sessionManager.Create(
		context.Background(),
		&auth.UserData{
			Login: nickname,
			Agent: r.UserAgent(),
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sess.Value,
		Expires:  time.Now().Add(expiration),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "VID",
		Value:    id,
		Expires:  time.Now().Add(expiration),
		HttpOnly: true,
	})
}

func getSession(r *http.Request) (*auth.UserData, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if err != nil {
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
	if _, err := getSession(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		return
	}

	////////////////////////
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

func uploadAvatar(r *http.Request) (urlAvatar string, err error) {
	cookieVID, err := r.Cookie("VID")
	if err != nil {
		return "", err
	}
	log.Println("DEBUG", "CookieVID", cookieVID.Value)
	err = r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		return "", err
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()
	log.Println("Molochnik")
	if _, err := os.Stat(pathToStaticFiles); os.IsNotExist(err) {
		err = os.Mkdir(pathToStaticFiles, 0700)
		if err != nil {
			return "", err
		}
	}
	dirname := cookieVID.Value
	if _, err := os.Stat(pathToStaticFiles + "/" + dirname); os.IsNotExist(err) {
		err = os.Mkdir(pathToStaticFiles+"/"+dirname, 0400)
		if err != nil {
			return "", err
		}
	}
	if err != nil {
		return "", err
	}

	hash := fnv.New64a()
	hash.Write([]byte(handler.Filename + time.Now().Format("15:04:05.00000")))
	hashname := string(hash.Sum64())
	fmt.Println("HASH:", hashname)

	saveFile, err := os.Create(pathToStaticFiles + "/" + dirname + "/" + hashname)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, file)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return hashname, nil
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	params := mux.Vars(r)
	userLogin := params["login"]
	fmt.Println("userLogin: ", userLogin)
	if userLogin == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	newData := updateDataImport{}
	newData.Email = r.FormValue("email")
	newData.Password = r.FormValue("password")
	newData.Nickname = r.FormValue("nickname")
	var url string
	if _, _, err := r.FormFile("file"); err != nil {
		log.Println("Field with this name not exist")
	} else {
		url, err = uploadAvatar(r)
		if err != nil {
			log.Printf("Upload Error: %T\n %s\n", err, err.Error())
			w.WriteHeader(http.StatusTeapot)
			return
		}
	}

	// ????
	// body, err := ioutil.ReadAll(r.Body)
	// defer r.Body.Close()
	// if err != nil {
	// 	log.Println(err)
	// }
	// if err := json.Unmarshal(body, &newData); err != nil {
	// 	log.Println(err)
	// }

	user := updateDataExport{}
	err := db.QueryRow(`
	UPDATE users SET
		email = COALESCE(NULLIF($1, ''), email),
		nickname = COALESCE(NULLIF($2, ''), nickname),
		password = COALESCE(NULLIF($3, ''), password),
		image = COALESCE(NULLIF($4, ''), image)
	WHERE nickname = $5
	AND  (NULLIF($1, '') IS NOT NULL AND NULLIF($1, '') IS DISTINCT FROM email OR
		 NULLIF($2, '') IS NOT NULL AND NULLIF($2, '') IS DISTINCT FROM nickname OR
		 NULLIF($3, '') IS NOT NULL AND NULLIF($3, '') IS DISTINCT FROM password OR
		 NULLIF($4, '') IS NOT NULL AND NULLIF($4, '') IS DISTINCT FROM password)
	RETURNING email, nickname, image;`,
		newData.Email, newData.Nickname, newData.Password, url, userLogin).Scan(&user.Email, &user.Nickname, &user.Image)
	switch err {
	case sql.ErrNoRows:
		log.Println("Method Update UserData: No rows were returned!")
		exportData := updateDataExport{
			Email:    newData.Email,
			Nickname: newData.Nickname,
			Image:    url,
		}
		json.NewEncoder(w).Encode(exportData)
	case nil:
		log.Println("new data of user: ", user)
		json.NewEncoder(w).Encode(user)
		return
	default:
		log.Println("Error updating record:", err)
		w.WriteHeader(http.StatusConflict)
		return
	}
}

// func mainPage(w http.ResponseWriter, r *http.Request) {
// 	session, err := getSession(r)
// 	if err != nil {
// 		log.Println("Error checking of session")
// 	}
// 	w.Write([]byte("WAO team"))

// 	if session != nil {
// 		w.Header().Set("Content-Type", "text/html")
// 		fmt.Fprintln(w, "\nWelcome "+session.Login)
// 	}
// }

func errorHandler(w http.ResponseWriter, r *http.Request, code int) {
	if code == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	http.Error(w, "", code)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().AddDate(0, -1, 0),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "VID",
		Value:    "",
		Expires:  time.Now().AddDate(0, -1, 0),
		HttpOnly: true,
	})

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

	// REPLACE
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

	row := db.QueryRow(`SELECT email, nickname, password FROM users WHERE nickname = $1 AND password = $2`, data.Nickname, data.Password)

	user := User{}
	switch err := row.Scan(&user.Email, &user.Nick, &user.password); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(`{"error": "Invalid login or password"}`))
	case nil:
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
	default:
		log.Println("Method Signin User: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Invalid login or password"}`))
		return
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	inputLogin := r.FormValue("login")

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
		Expires: time.Now().Add(expiration),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
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
	apiV1.HandleFunc("/users", CreateUser).Methods("POST", "OPTIONS")
	apiV1.HandleFunc("/users/{login}", GetUser).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/users/man/{login}", UpdateUser) //.Methods("PUT", "OPTIONS")
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
	// http.HandleFunc("/login", loginPage)
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
				Image: <input type="file" name="file">
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

	srv := &http.Server{
		Handler:      siteMux,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server at http://" + srv.Addr)
	log.Println(srv.ListenAndServe())
}
