package main

import (
	"net/http"
	"strconv"
	"os"
	"log"
	"context"
	"os/signal"
	"syscall"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"github.com/DmitriyPrischep/backend-WAO/pkg/game"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	GameController *game.Game
	sessionManager auth.AuthCheckerClient
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
		log.Println("Socket connection error", err)
		return
	}
	player := game.NewPlayer(ws)
	GameController.AddPlayer(player)
}

func main() {
	viper.SetConfigFile(os.Args[1])
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}
	port := viper.GetString("game.port")
	hostAuth := viper.GetString("authsrv.host") + ":" + viper.GetString("authsrv.port")

	grpcConnect, err := grpc.Dial(
		hostAuth,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("Can't connect to gRPC")
	}
	sessionManager = auth.NewAuthCheckerClient(grpcConnect)

	router := mux.NewRouter()
	router.HandleFunc("/websocket", SocketFunc)
	http.Handle("/", router)
	
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.Printf("caught sig: %+v", sig)
		grpcConnect.Close()
		log.Println("Connections close")
		os.Exit(0)
	}()

	// viper.SetConfigFile("./config/env.yml")
	// err := viper.ReadInConfig()
	// if err != nil {
	// 	panic(err)
	// }
	countRoom := viper.GetString("game.countRoom")
	cnt, err := strconv.ParseUint(countRoom, 10, 32)
	if err != nil {
		log.Printf("Type: %T %v\n", cnt, cnt)
	}
	GameController = game.NewGame(uint(cnt))
	go GameController.Run()
	log.Println("Server is listening")
	http.ListenAndServe(":"+port, nil)
}
