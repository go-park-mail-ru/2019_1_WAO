package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type User struct {
	ID       int    `json:"id, string, omitempty"`
	Email    string `json:"email, omitempty"`
	password string `json:"password, omitempty"`
	Nick     string `json:"nick, omitempty"`
	Scope    int    `json:"scope, string, omitempty"`
	Games    int    `json:"games, string, omitempty"`
	Wins     int    `json:"wins, string, omitempty"`
	// Image
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
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r) // Get Params
	for _, item := range users {

		userID, err := strconv.Atoi(params["id"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if item.ID == userID {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	http.Error(w, `{"error": "This user is not found}"`, http.StatusNotFound)
}

//Create New User
func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	user.ID = len(users)
	users = append(users, user)
	json.NewEncoder(w).Encode(user)
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

func main() {
	r := mux.NewRouter()

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
		},
		User{
			ID:       2,
			Email:    "pashok@pochta.ru",
			password: "12345",
			Nick:     "joker",
			Scope:    200,
			Games:    1,
			Wins:     3,
		},
		User{
			ID:       3,
			Email:    "karman@pochta.ru",
			password: "12345",
			Nick:     "gopher",
			Scope:    88,
			Games:    8,
			Wins:     0,
		},
		User{
			ID:       4,
			Email:    "support@pochta.ru",
			password: "12345",
			Nick:     "Batman",
			Scope:    13,
			Games:    11,
			Wins:     0,
		},
	)

	r.HandleFunc("/api/users", getUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/api/users", createUser).Methods("POST")
	r.HandleFunc("/api/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", deleteUser).Methods("DELETE")
	fmt.Println("Starting server at http://127.0.0.1:8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
