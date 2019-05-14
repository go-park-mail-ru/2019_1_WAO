package main

import (
	"fmt"
	"net/http"
	"log"
	"context"

	"github.com/DmitriyPrischep/backend-WAO/pkg/game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var (
	sessionManager auth.AuthCheckerClient
	GameController *game.Game
)


func SocketFunc(w http.ResponseWriter, r *http.Request) {

	cookieSessionID, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = sessionManager.Check(
		context.Background(),
		&auth.Token{
			Value: cookieSessionID.Value,
		})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Socket connection error", err)
		return
	}

	player := game.NewPlayer(ws)
	GameController.AddPlayer(player)
}


func main() {
	viper.AddConfigPath("../../")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}
	port := viper.GetString("game.port")
	statics := viper.GetString("game.static")
	hostAuth := viper.GetString("authsrv.host") + ":" + viper.GetString("authsrv.port")

	grpcConnect, err := grpc.Dial(
		hostAuth,
		grpc.WithInsecure(),
	)

	if err != nil {
		log.Println("Can't connect to gRPC")
	}
	defer grpcConnect.Close()

	sessionManager = auth.NewAuthCheckerClient(grpcConnect)

	router := mux.NewRouter()
	router.HandleFunc("/websocket", SocketFunc)
	http.Handle("/", router)
	fs := http.FileServer(http.Dir(statics))
	http.Handle("/"+statics+"/", http.StripPrefix("/"+statics, fs))
	fmt.Println("Server is listening")
	GameController = game.NewGame(1) // New GameController
	go GameController.Run()
	http.ListenAndServe(":" + port, nil)
}
