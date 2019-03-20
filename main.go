package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
)

var signInFormTmpl = []byte(`
<html>
	<body>
		<form action="/v1/signin" method="post">
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
		<form action="/v1/users" method="post" enctype="multipart/form-data">
			Email:<input type="text" name="email">
			Password:<input type="password" name="password">
			Nickname:<input type="text" name="nickname"><br>
			Avatar: <input type="file" name="input_file">
			<input type="submit" value="Register">
		</form>
	</body>
</html>
`)

var SECRET = []byte("secretkey")

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

//Init users var as a slise User struct
var users []User

func checkAuthorization(r http.Request) (bool, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return false, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, err
		}
		return SECRET, nil
	})
	if err != nil {
		log.Printf("Unexpected signing method: %v", token.Header["alg"])
		return false, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	}
	return false, nil
}

//Get All Users
func GetUsers(w http.ResponseWriter, r *http.Request) {
	if ok, err := checkAuthorization(*r); !ok {
		log.Println("Autorization checking error:", err)
		http.Redirect(w, r, "/v1/login", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
	// io.WriteString(w, `{"status": 500, "err": "db_error"}`)
	// w.WriteHeader(http.StatusOK)
}

//Get Single User
func getUser(w http.ResponseWriter, r *http.Request) {
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

	if ok, err := checkAuthorization(*r); !ok {
		log.Println("Autorization checking error:", err)
		http.Redirect(w, r, "/v1/login", http.StatusSeeOther)
	}

	w.Header().Set("Content-Type", "text/html")

	params := mux.Vars(r)
	for _, item := range users {
		userID, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if item.ID == userID {
			w.Write([]byte("Player profile "))
			w.Write([]byte(item.Nick + "<br>"))
			urlImage := fmt.Sprintf(`<img src="/data/%d/%s"/>`, item.ID, item.Image)
			w.Write([]byte(urlImage))
			return
		}
	}
}

//Create New User
func createUser(w http.ResponseWriter, r *http.Request) {
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
	user.ID = len(users) + 1
	user.Email = r.FormValue("email")
	user.password = r.FormValue("password")
	user.Nick = r.FormValue("nickname")
	user.Scope = 0
	user.Wins = 0
	user.Games = 0

	err := r.ParseMultipartForm(5 * 1024 * 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	file, handler, err := r.FormFile("input_file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

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
	users = append(users, user)
	http.Redirect(w, r, "/v1/", http.StatusSeeOther)
}

//Update the User
func updateUser(w http.ResponseWriter, r *http.Request) {
	if ok, err := checkAuthorization(*r); !ok {
		log.Println("Autorization checking error:", err)
		http.Redirect(w, r, "/v1/login", http.StatusSeeOther)
	}

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	for i, item := range users {
		userID, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if item.ID == userID {
			var user User
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			users[i].Email = user.Email
			users[i].password = user.password
			users[i].Nick = user.Nick
			break
		}
	}
	w.WriteHeader(http.StatusOK)
}

//Delete the User
func deleteUser(w http.ResponseWriter, r *http.Request) {
	if ok, err := checkAuthorization(*r); !ok {
		log.Println("Autorization checking error:", err)
		http.Redirect(w, r, "/v1/login", http.StatusSeeOther)
	}

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range users {
		userID, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if item.ID == userID {
			users = append(users[:index], users[index+1:]...)
			break
		}
	}
	w.WriteHeader(http.StatusOK)
}

func signup(w http.ResponseWriter, r *http.Request) {
	w.Write(signUpFormTmpl)
}

func signin(w http.ResponseWriter, r *http.Request) {
	var userExist bool
	for _, val := range users {
		if val.password == r.FormValue("password") && val.Email == r.FormValue("login") {
			userExist = true
			break
		}
	}

	if !userExist {
		http.Redirect(w, r, "/v1/signup", http.StatusSeeOther)
		return
	}

	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": r.FormValue("login"),
		"exp":      time.Now().Add(1 * time.Minute).Unix(),
	})

	token, err := rawToken.SignedString(SECRET)
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
	http.Redirect(w, r, "/v1/", http.StatusSeeOther)
	return
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WAO team"))
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Write(signInFormTmpl)
}

func logout(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
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

// Mock Data - implement DB
func MockDB() {
	users = append(users,
		User{
			ID:       1,
			Email:    "goshan@pochta.ru",
			password: "12345",
			Nick:     "karlik",
			Scope:    119,
			Games:    5,
			Wins:     1,
			Image:    "avatar.jpg",
		},
		User{
			ID:       2,
			Email:    "pashok@pochta.ru",
			password: "12345",
			Nick:     "joker",
			Scope:    200,
			Games:    1,
			Wins:     3,
			Image:    "avatar.jpg",
		},
		User{
			ID:       3,
			Email:    "karman@pochta.ru",
			password: "12345",
			Nick:     "gopher",
			Scope:    88,
			Games:    8,
			Wins:     0,
			Image:    "avatar.jpg",
		},
		User{
			ID:       4,
			Email:    "support@pochta.ru",
			password: "12345",
			Nick:     "Batman",
			Scope:    13,
			Games:    11,
			Wins:     0,
			Image:    "avatar.jpg",
		},
	)
}

func main() {
	MockDB()

	r := mux.NewRouter()
	r.HandleFunc("/", redirectOnMain).Methods("GET")

	v1 := r.PathPrefix("/v1").Subrouter()

	v1.HandleFunc("/users", GetUsers).Methods("GET")
	v1.HandleFunc("/users/{id}", getUser).Methods("GET")
	v1.HandleFunc("/users", createUser).Methods("POST")
	v1.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	v1.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	v1.HandleFunc("/", mainPage)
	v1.HandleFunc("/signup", signup).Methods("GET")
	v1.HandleFunc("/signin", signin).Methods("POST")
	v1.HandleFunc("/login", login).Methods("GET")
	v1.HandleFunc("/logout", logout).Methods("GET")

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("./static/"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server at http://127.0.0.1:8000/")
	log.Println(srv.ListenAndServe())
}
