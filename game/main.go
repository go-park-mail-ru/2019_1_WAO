package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	game := NewGame(10)
	go game.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{}

		cookie, err := r.Cookie("auth")
		if err != nil {
			log.Println("not authorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, http.Header{"Upgrade": []string{"websocket"}})
		if err != nil {
			log.Printf("error while connecting: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("connected to client")
		player := NewPlayer(conn, cookie.Value)
		go player.Listen()
		game.AddPlayer(player)
	})

	http.ListenAndServe(":8080", nil)
}
