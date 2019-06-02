package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"os/signal"
	"syscall"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/aws"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/router"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	sessionManager auth.AuthCheckerClient
)

const (
	expiration = 360 * time.Minute
)

func main() {
	if len(os.Args) > 1 {
		viper.SetConfigFile(os.Args[1])
	} else {
		log.Println("Config file is not select")
		return
	}
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Cannot read config", err)
		return
	}

	userDB := viper.GetString("db.user")
	userPass := viper.GetString("db.password")
	nameDB := viper.GetString("db.name")
	sslMode := viper.GetString("db.sslmode")
	connection, err := driver.ConnectSQL(userDB, userPass, nameDB, sslMode)
	if err != nil {
		fmt.Println(err)
		return
	}

	hostAuth := viper.GetString("authsrv.host") + ":" + viper.GetString("authsrv.port")
	grpcConnect, err := grpc.Dial(
		hostAuth,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("Can't connect to gRPC")
	}

	sessionManager = auth.NewAuthCheckerClient(grpcConnect)

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		log.Printf("caught sig: %+v", sig)
		connection.DB.Close()
		grpcConnect.Close()
		log.Println("Connections close")
		os.Exit(0)
	}()

	port := viper.GetString("apisrv.port")
	host := viper.GetString("apisrv.host")

	setting := &aws.ConnectSetting{
		AccessKeyID:     viper.GetString("aws.keyID"),
		SecretAccessKey: viper.GetString("aws.secretKey"),
		Token:           viper.GetString(""),
		Region:          viper.GetString("aws.region"),
		NameBucket:      viper.GetString("aws.bucket"),
		PathRootDir:     viper.GetString("aws.root"),
	}

	api := viper.GetString("apisrv.api")
	urlCORS := viper.GetString("apisrv.urlCORS")
	urlImage := viper.GetString("apisrv.urlImage")

	router := router.CreateRouter(api, urlCORS, urlImage, sessionManager, connection, setting)

	srv := &http.Server{
		Handler:      router,
		Addr:         host + ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("Starting server at http://" + srv.Addr)
	log.Println(srv.ListenAndServe())
}