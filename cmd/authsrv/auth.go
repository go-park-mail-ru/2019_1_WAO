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


	connectStr := "user=postgres password=123456 dbname=waogame sslmode=disable"
	// connectStr2 := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	// 	viper.GetString("db.user"), viper.GetString("db.password"), viper.GetString("db.name"))
	// fmt.Printf("Equal:\n%s\n%s", connectStr, connectStr2)

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
	userDB := viper.GetString("db.user")
	userPass := viper.GetString("db.password")
	nameDB := viper.GetString("db.name")
	sslMode := viper.GetString("db.sslmode")

	connectStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
		userDB, userPass, nameDB, sslMode)
	var err error
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
	}
	
	defer db.Close()
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
