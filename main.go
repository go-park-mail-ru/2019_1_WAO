package main

import (
	"fmt"
	"log"
	"net/http"

	"./game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func MoveFunc(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("Error", err)
	}
	// defer ws.Close()
	go ReadSocket(ws)
}

func ReadSocket(ws *websocket.Conn) {
	fmt.Println("TEST")
	for {
		// _, message, err := ws.ReadMessage()
		type message struct {
			Id       int    `json:"id"`
			Name     string `json:"name"`
			EmpField string
		}
		var msg message
		err := websocket.ReadJSON(ws, &msg)
		if err != nil {
			fmt.Println("Disconnect")
			// ws.Close()
			break
		}
		fmt.Println("id - ", msg.Id)
		fmt.Println("name = ", msg.Name)
		msg.Id += 3
		msg.Name = "Anonymous"
		if err := websocket.WriteJSON(ws, msg); err != nil {
			fmt.Println("Disconnect")
			break
		}
		// fmt.Println(message)
		// fmt.Println("***", string(message), "***")
		// ws.WriteMessage(websocket.TextMessage, message)
	}
}

func WriteSocket(ws *websocket.Conn) {
	// ticker := time.NewTicker(3 * time.Second)
	// for {
	// 	w, err := ws.NextWriter(websocket.TextMessage)
	// 	if err != nil {
	// 		ticker.Stop()
	// 		break
	// 	}
	// 	w.Write([]byte("Hello!"))
	// 	w.Close()
	// 	<-ticker.C
	// }
	player := game.FieldGenerator(100, 100, 10.0)

	err := ws.WriteJSON(player)
	if err != nil {
		fmt.Println("Error was occured", err)
	}
}
func MainFunc(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// id := vars["id"]
	// tmpl, _ := template.ParseFiles("./templates/index.html")
	// tmpl.Execute(w, "")
	http.ServeFile(w, r, "index.html")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", MainFunc)
	router.HandleFunc("/move", MoveFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")

	http.ListenAndServe(":8080", nil)
}
