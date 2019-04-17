package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
		_, message, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Disconnect")
			// ws.Close()
			break
		}
		fmt.Println(message)
		fmt.Println("***", string(message), "***")
		ws.WriteMessage(websocket.TextMessage, message)
		// if err != nil {
		// 	log.Fatal(err)
		// 	break
		// }
	}
	// ws.Close()
	// log.Println("Message:")
	// fmt.Println(result)
	// w, err := ws.NextWriter(websocket.TextMessage)
	// if err != nil {
	// 	log.Fatal("Error", err)
	// 	return
	// }
	// myMessage := []byte("Hello, World!")
	// fmt.Println(myMessage)
	// w.Write(myMessage)
	// log.Println("Message was sent")
	// w.Close()
}

func WriteSocket(ws *websocket.Conn) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		w, err := ws.NextWriter(websocket.TextMessage)
		if err != nil {
			ticker.Stop()
			break
		}
		w.Write([]byte("Hello!"))
		w.Close()
		<-ticker.C
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
	// http.Handle("/socket.io/", server)

	http.ListenAndServe(":8080", nil)
}
