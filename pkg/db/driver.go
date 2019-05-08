package db

import (
	"database/sql"
	"log"
	"io/ioutil"
	"fmt"
)

type BDService struct {
	DB *sql.DB
}

// type DBService interface {
// 	User(id int) (*User, error)
// 	Users() ([]*User, error)
// 	CreateUser(u *User) error
// 	DeleteUser(id int) error
// }

// User returns a user for a given id.
func (s *DBService) User(id int) (*User, error) {
	var u myapp.User
	row := db.QueryRow(`SELECT id, name FROM users WHERE id = $1`, id)
	if row.Scan(&u.ID, &u.Name); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *DBService) Users() ([]*User, error){
	rows, err := db.Query("SELECT id, email, nickname, scope, games, wins, image FROM users ORDER by scope DESC;")
	defer rows.Close()
	if err != nil {
		// log.Println("Method GetUsers:", err)
		return nil, err
	}
	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Nick, &user.Score, &user.Games, &user.Wins, &user.Image)
		if err != nil {
			log.Println(err)
			continue
		}
		if user.Image != "" {
			user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
		}
		users = append(users, user)
	}
	return users, nil
}

func connectDB(){
	// connStr := "user=postgres password=123456 dbname=waogame sslmode=disable"
	connectStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	var err error
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Printf("No connection to DB: %v", err)
	}
}


func DumpDB(db *sql.DB) error {
	buffer, err := ioutil.ReadFile(schema)
	if err != nil {
		return err
	}

	schema := string(buffer)
	_, err = db.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}


rows, err := db.Query("SELECT id, email, nickname, scope, games, wins, image FROM users ORDER by scope DESC;")
	if err != nil {
		log.Println("Method GetUsers:", err)
	}
	defer rows.Close()
	users := []User{}

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Nick, &user.Score, &user.Games, &user.Wins, &user.Image)
		if err != nil {
			log.Println(err)
			continue
		}
		if user.Image != "" {
			user.Image = fmt.Sprintf(`/data/%d/%s`, user.ID, user.Image)
		}
		users = append(users, user)
	}