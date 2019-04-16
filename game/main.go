package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	game := NewGame(10)
	go game.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{}

		resp, err := http.Get("https://127.0.0.7:8000/api/session")
		if err != nil {
			fmt.Println(err)
			return
		}

		b, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg := AuthResponce{}
		err = json.Unmarshal(b, &msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// cookie, err := r.Cookie("auth")
		// if err != nil {
		// 	log.Println("not authorized")
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		conn, err := upgrader.Upgrade(w, r, http.Header{"Upgrade": []string{"websocket"}})
		if err != nil {
			log.Printf("error while connecting: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("connected to client")
		player := NewPlayer(conn, msg.nickname)
		go player.Listen()
		game.AddPlayer(player)
	})

	http.ListenAndServe(":8080", nil)
}
