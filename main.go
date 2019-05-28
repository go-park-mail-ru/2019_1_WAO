package main

import (
	"fmt"
	"net/http"

	game "./game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var GameController *game.Game

func SocketFunc(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Socket connection error", err)
		return
	}

	// !!!
	player := game.NewPlayer(ws)
	GameController.AddPlayer(player)
}

func MainFunc(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// tmpl, _ := template.ParseFiles("./templates/index.html")
	// tmpl.Execute(w, "")

	// fmt.Fprintf(w, "ok")
	// game.GameLoop() // Init game process
	http.ServeFile(w, r, "index.html")

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", MainFunc)
	router.HandleFunc("/websocket", SocketFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")
	// Rooms = make(map[string]connections)
	viper.SetConfigFile("./config/env.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	GameController = game.NewGame(5) // New GameController
	go GameController.Run()
	http.ListenAndServe(":8080", nil)
}
