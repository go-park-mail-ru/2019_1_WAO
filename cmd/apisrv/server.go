package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/aws"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/router"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	sessionManager auth.AuthCheckerClient
	db             *sql.DB
)

const (
	expiration = 10 * time.Minute
)

func main() {
	viper.SetConfigFile(os.Args[1])
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
	defer grpcConnect.Close()

	sessionManager = auth.NewAuthCheckerClient(grpcConnect)

	defer db.Close()

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
