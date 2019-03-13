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

	"github.com/gorilla/mux"
)

var signInFormTmpl = []byte(`
<html>
	<body>
		<form action="/v1/" method="post">
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

//Get All Users
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
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
	w.Header().Set("Content-Type", "text/html")

	params := mux.Vars(r) // Get Params
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
	log.Println("Method POST add user")
	if r.Method != http.MethodPost {
		w.Write(signUpFormTmpl)
		return
	}
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
			log.Fatal(err.Error())
		}
	}
	if err != nil {
		log.Fatal(err.Error())
	}

	saveFile, err := os.Create("./static/" + dirname + "/" + handler.Filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer saveFile.Close()

	_, err = io.Copy(saveFile, file)
	if err != nil {
		log.Fatal(err)
	}

	user.Image = handler.Filename
	users = append(users, user)
	http.Redirect(w, r, "/v1/", http.StatusSeeOther)
}

//Update the User
func updateUser(w http.ResponseWriter, r *http.Request) {
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
	w.Write(signInFormTmpl)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("WAO team"))
}

func main() {

	// Mock Data - implement DB
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

	r := mux.NewRouter()
	v1 := r.PathPrefix("/v1").Subrouter()

	v1.HandleFunc("/users", getUsers).Methods("GET")
	v1.HandleFunc("/users/{id}", getUser).Methods("GET")
	v1.HandleFunc("/users", createUser).Methods("POST")
	v1.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	v1.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	v1.HandleFunc("/", mainPage)
	v1.HandleFunc("/signup", signup).Methods("GET")
	v1.HandleFunc("/signin", signin).Methods("GET")

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("./static/"))))

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("Starting server at http://127.0.0.1:8000")
	log.Fatal(srv.ListenAndServe())
}
