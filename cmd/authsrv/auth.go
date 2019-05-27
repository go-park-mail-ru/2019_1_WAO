package main 

import (
	"fmt"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"log"
	"net"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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
		log.Fatalln("cant listet port", err)
	}
	
	server := grpc.NewServer()
	auth.RegisterAuthCheckerServer(server, NewSessionManager())
	fmt.Println("Auth Service starting server at http://" + host + ":" + port)
	server.Serve(listener)
}
