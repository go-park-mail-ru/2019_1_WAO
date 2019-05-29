package main 

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
)

var (
	secret string
)

func main() {
	viper.SetConfigFile(os.Args[1])
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}
	secret = viper.GetString("secretkey")
	
	port := viper.GetString("authsrv.port")
	host := viper.GetString("authsrv.host")
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalln("Can't listet port", err)
	}

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.Printf("Caught sig: %+v\n Graceful stop service", sig)
		os.Exit(0)
	}()
	
	server := grpc.NewServer()
	auth.RegisterAuthCheckerServer(server, NewSessionManager())
	fmt.Println("Auth Service starting server at http://" + host + ":" + port)
	server.Serve(listener)
}
