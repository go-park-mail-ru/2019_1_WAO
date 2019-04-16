package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	socketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func SocketFunc(w http.ResponseWriter, r *http.Request) {
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
	router.HandleFunc("/echo", SocketFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	fmt.Println("Server is listening")
	server := socketio.NewServer(transport.GetDefaultWebsocketTransport())
	server.On(socketio.OnConnection, func(s *socketio.Channel) {

		log.Println("***Connection***")
		s.Join("players")
	})
	server.On("left", func(s *socketio.Channel, msg string) {
		// var data map[string]interface{}
		// if err := json.Unmarshal(msg, &data); err != nil {
		// 	log.Fatal("Error", err)
		// }
		// log.Println(data)
		log.Println("Left key pressed (x -= 3)")
		// so.BroadcastTo("chat", "chat message", msg)
	})
	server.On("right", func(so *socketio.Channel, msg string) {
		log.Println("Right key pressed (x += 3)")
		// so.BroadcastTo("chat", "chat message", msg)
	})
	server.On(socketio.OnDisconnection, func(s *socketio.Channel) {
		s.Leave("players")
		log.Println("Disconnect...")
	})
	server.On("error", func(s *socketio.Channel, err error) {
		log.Println("error:", err)
	})

	http.Handle("/socket.io/", server)

	http.ListenAndServe(":8080", nil)
}
