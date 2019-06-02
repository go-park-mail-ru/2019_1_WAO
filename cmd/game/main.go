package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	// game "./game"
	// "github.com/go-park-mail-ru/2019_1_WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/game"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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
		log.Println("BAD REQUEST")
		return
	}
	sess, err := sessionManager.Check(
		context.Background(),
		&auth.Token{
			Value: cookieSessionID.Value,
		})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("UNAUTHORIZED")
		return
	}
	// id := sess.Id
	id, err := strconv.Atoi(sess.Id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("No ID")
		return
	}
	// nickname := sess.Login

	log.Println("DATA:", id)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Socket connection error", err)
		return
	}
	// !!!
	player := game.NewPlayer(ws, id)
	GameController.AddPlayer(player)
}

func main() {
	viper.SetConfigFile(os.Args[1])
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}
	port := viper.GetString("gamesrv.port")
	hostAuth := viper.GetString("authsrv.host") + ":" + viper.GetString("authsrv.port")

	grpcConnect, err := grpc.Dial(
		hostAuth,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("Can't connect to gRPC")
		return
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
	fmt.Println("Server is listening")
	countRoom := viper.GetString("game.countRoom")
	cnt, err := strconv.ParseUint(countRoom, 10, 32)
	if err != nil {
		log.Printf("Type: %T %v\n", cnt, cnt)
	}
	GameController = game.NewGame(uint(cnt)) // New GameController
	go GameController.Run()
	http.ListenAndServe(":"+port, nil)
}
