package driver

import (
	"log"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
)


type DB struct {
	DB *sql.DB
}

var dbConnect = &DB{}

func ConnectSQL(uname, pass, dbname, sslmode string) (*DB, error) {

	connectStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
		uname, pass, dbname, sslmode)
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
		return nil, err
	}
	dbConnect.DB = db
	return dbConnect, err
}