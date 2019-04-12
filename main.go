package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	// "os"
	// "strconv"
	"time"
	"io/ioutil"

	// _ "github.com/DmitriyPrischep/backend-WAO/docs"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/gorilla/handlers"
)

var signInFormTmpl = []byte(`
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

var signUpFormTmpl = []byte(`
<html>
	<body>
		<form action="/api/users" method="post" enctype="multipart/form-data">
			Email:<input type="text" name="email">
			Password:<input type="password" name="password">
			Nickname:<input type="text" name="nickname"><br>
			Avatar: <input type="file" name="input_file">
			<input type="submit" value="Register">
		</form>
	</body>
</html>
`)

var secret = []byte("secretkey")

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

//Init users var as a slise User struct
var Users []User

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Users)
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
	// w.Header().Set("Content-Type", "application/json")
	// params := mux.Vars(r) // Get Params
	// for _, item := range users {
	// 	userID, err := strconv.Atoi(params["id"])
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusNotFound)
	// 	}
	// 	if item.ID == userID {
	// 		json.NewEncoder(w).Encode(item)
	// 		return
	// 	}
	// }
	// http.Error(w, `{"error": "This user is not found}"`, http.StatusNotFound)

	w.Header().Set("Content-Type", "text/html")

	params := mux.Vars(r)
	for _, item := range Users {
		userLogin := params["login"]
		if item.Nick == userLogin {
			w.Write([]byte("Player profile "))
			w.Write([]byte(item.Nick + "<br>"))
			urlImage := fmt.Sprintf(`<img src="/data/%d/%s"/>`, item.ID, item.Image)
			w.Write([]byte(urlImage))
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}

type SignupData struct {
	Nick string `json:"nickname,omitempty"`
	Email string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
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
	// w.Header().Set("Content-Type", "application/json")
	// var user User
	// err := json.NewDecoder(r.Body).Decode(&user)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// }
	// user.ID = len(users)
	// users = append(users, user)
	// json.NewEncoder(w).Encode(user)

	var user User
	user.ID = len(Users) + 1
	// user.Email = r.FormValue("email") // так не распарсит
	// user.password = r.FormValue("password")
	// user.Nick = r.FormValue("nickname")
	body, err := ioutil.ReadAll(r.Body)
	fmt.Println(body)
	defer r.Body.Close()
	signupData := &SignupData{}
	err = json.Unmarshal(body, &signupData)
	fmt.Println(signupData)
	fmt.Println(err)

	user.Email = signupData.Email
	user.password = signupData.Password
	user.Nick = signupData.Nick
	user.Scope = 0
	user.Wins = 0
	user.Games = 0

	if user.Email == "" || user.password == "" {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"error": "Invalid data"}`)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// при создании пользователя картинка не загружается
	// а выставляется дефолтная
	/* 
	err = r.ParseMultipartForm(5 * 1024 * 1024) // err :=
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
	*/
	Users = append(Users, user)
	// http.Redirect(w, r, "/", http.StatusOK)
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Nick,
		"exp":      time.Now().Add(1 * time.Minute).Unix(),
	})
	token, err := rawToken.SignedString(secret)
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    token,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
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

	for i, item := range Users {
		if item.Nick == userLogin {
			var user User
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			Users[i].Email = user.Email
			Users[i].password = user.password
			Users[i].Nick = user.Nick
			break
		}
	}
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
	for index, item := range Users {
		userLogin := params["login"]
		if userLogin == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if item.Nick == userLogin {
			Users = append(Users[:index], Users[index+1:]...)
			break
		}
	}
	w.WriteHeader(http.StatusOK)
}

func signup(w http.ResponseWriter, r *http.Request) {
	w.Write(signUpFormTmpl)
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

	var userExist bool
	for _, val := range Users {
		if val.password == r.FormValue("password") && val.Email == r.FormValue("login") {
			userExist = true
			break
		}
	}

	if !userExist {
		// http.Redirect(w, r, "/signup", http.StatusSeeOther) // все редиректы на фронте
		w.WriteHeader(http.StatusSeeOther)
		return
	}

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
	w.WriteHeader(http.StatusOK)
}

func redirectOnMain(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/api/", http.StatusSeeOther)
	return
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WAO team"))
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Write(signInFormTmpl)
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

	email, ok := claims["username"]
	if !ok {
		log.Println("Bad claims: field 'username' not exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	for _, item := range Users {
		if item.Email == email {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
}

// Mock Data - implement DB
func MockDB() {
	Users = append(Users,
		User{
			ID:       1,
			Email:    "goshan@pochta.ru",
			password: "123456",
			Nick:     "karlik",
			Scope:    119,
			Games:    5,
			Wins:     1,
			Image:    "avatar.jpg",
		},
		User{
			ID:       2,
			Email:    "pashok@pochta.ru",
			password: "123456",
			Nick:     "joker",
			Scope:    200,
			Games:    1,
			Wins:     3,
			Image:    "avatar.jpg",
		},
		User{
			ID:       3,
			Email:    "karman@pochta.ru",
			password: "123456",
			Nick:     "gopher",
			Scope:    88,
			Games:    8,
			Wins:     0,
			Image:    "avatar.jpg",
		},
		User{
			ID:       4,
			Email:    "support@pochta.ru",
			password: "123456",
			Nick:     "Batman",
			Scope:    13,
			Games:    11,
			Wins:     0,
			Image:    "avatar.jpg",
		},
	)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok, err := checkAuthorization(*r); !ok {
			log.Println(err.Error())
			// http.Redirect(w, r, "/login", http.StatusTemporaryRedirect) // все редиректы на фронте
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("accessLogMiddleware", r.URL.Path)
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
	MockDB()

	actionMux := mux.NewRouter()

	apiV1 := actionMux.PathPrefix("/api").Subrouter()
	apiV1.HandleFunc("/users", GetUsers).Methods("GET")
	apiV1.HandleFunc("/users/{login}", GetUser).Methods("GET")
	apiV1.HandleFunc("/users", CreateUser).Methods("POST")
	apiV1.HandleFunc("/users/{login}", updateUser).Methods("PUT")
	apiV1.HandleFunc("/users/{login}", deleteUser).Methods("DELETE")
	apiV1.HandleFunc("/session", checkSession).Methods("GET")
	apiV1.HandleFunc("/session", logout).Methods("DELETE")
	apiV1.HandleFunc("/signup", CreateUser)
	apiV1.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Ппссс, парень! Такой страницы не существует!"))
	})

	// apiV1.Use(authMiddleware)

	siteMux := http.NewServeMux()
	siteMux.Handle("/api/", apiV1)
	siteMux.HandleFunc("/docs/", httpSwagger.WrapHandler)
	siteMux.HandleFunc("/", mainPage)
	// siteMux.HandleFunc("/signup", signup)
	// siteMux.HandleFunc("/signup", CreateUser)
	siteMux.HandleFunc("/signin", signin)
	siteMux.HandleFunc("/login", login)
	siteMux.HandleFunc("/logout", logout)
	siteMux.Handle("/favicon.ico", http.NotFoundHandler())

	staticHandler := http.StripPrefix(
		"/data/",
		http.FileServer(http.Dir("./static")),
	)
	siteMux.Handle("/data/", staticHandler)

	siteHandler := accessLogMiddleware(siteMux)
	siteHandler = panicMiddleware(siteHandler)

	/*
	srv := &http.Server{
		Handler:      siteHandler,
		Addr:         "127.0.0.1",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	*/

	// CORS
	// frontAdr := "https://wao2019.herokuapp.com/"
	frontAdr := "http://127.0.0.1:3000"
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Origin"})
	originsOk := handlers.AllowedOrigins([]string{frontAdr})
	crendenOk := handlers.AllowCredentials()
	methodsOk := handlers.AllowedMethods([]string{"GET", "DELETE", "POST", "PUT", "OPTIONS"})

	log.Println("Starting server at http://127.0.0.1:8000/")
	log.Println(http.ListenAndServe(":8000", handlers.CORS(originsOk, crendenOk, headersOk, methodsOk)(actionMux)))
}
