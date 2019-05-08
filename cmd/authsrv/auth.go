package main 

import (
	"fmt"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"net"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var db *sql.DB

func init() {
	connectStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("db.user"), viper.GetString("db.password"), viper.GetString("db.name"))
	var err error
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
	}
}

func main() {
	viper.AddConfigPath("../../")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}
	defer db.Close()
	port := viper.GetString("authsrv.port")
	listener, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer()

	auth.RegisterAuthCheckerServer(server, NewSessionManager())

	fmt.Println("Auth Service starting server at http://" + viper.GetString("authsrv.host") + ":" + port)
	server.Serve(listener)
}
