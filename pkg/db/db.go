package db

import (
	"database/sql"
	"log"
	"fmt"
	"github.com/DmitriyPrischep/backend-WAO/pkg/model"
	"github.com/DmitriyPrischep/backend-WAO/pkg/methods"
)

func NewDataBase(conn *sql.DB) methods.UserMethods {
	return &DBService{
		DB: conn,
	}
}

type DBService struct {
	DB *sql.DB
}

func (s *DBService) GetUsers() (users []model.User, err error) {
	rows, err := s.DB.Query("SELECT id, email, nickname, scope, games, wins, image FROM users ORDER by scope DESC;")
	if err != nil {
		log.Println("Method GetUsers:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := model.User{}
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
	return users, err
}
func (s *DBService) GetUser(userdata model.NicknameUser) (user *model.User, err error) {
	row := s.DB.QueryRow(`SELECT id, email, nickname, scope, games, wins, image 
						FROM users WHERE nickname = $1;`, userdata.Nickname)

	switch err := row.Scan(&user.ID, &user.Email, &user.Nick, &user.Score,
		&user.Games, &user.Wins, &user.Image); err {
	case sql.ErrNoRows:
		log.Println("Method GetUser: No rows were returned!")
		return nil, err
	case nil:
		return user, err
	default:
		log.Printf("Error type: %T: Method GetUser: %s\n", err, err.Error())
		return nil, err
	}
}

func (s *DBService) CreateUser(user model.UserRegister) (nickname string, err error) {
	err = s.DB.QueryRow(`INSERT INTO users (email, nickname, password, scope, games, wins, image)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING nickname`,
	user.Email, user.Nickname, user.Password, 0, 0, 0, "").Scan(&nickname)
	if err != nil {
		log.Printf("Error inserting record: %v", err)
		return "", err
	}
	return user.Nickname, nil
}

func (s *DBService) UpdateUser(user model.UpdateDataImport) (out model.UpdateDataExport, err error) {
	err = s.DB.QueryRow(`
	UPDATE users SET
		email = COALESCE(NULLIF($1, ''), email),
		nickname = COALESCE(NULLIF($2, ''), nickname),
		password = COALESCE(NULLIF($3, ''), password),
		image = COALESCE(NULLIF($4, ''), image)
	WHERE nickname = $5
	AND  (NULLIF($1, '') IS NOT NULL AND NULLIF($1, '') IS DISTINCT FROM email OR
		 NULLIF($2, '') IS NOT NULL AND NULLIF($2, '') IS DISTINCT FROM nickname OR
		 NULLIF($3, '') IS NOT NULL AND NULLIF($3, '') IS DISTINCT FROM password OR
		 NULLIF($4, '') IS NOT NULL AND NULLIF($4, '') IS DISTINCT FROM password)
	RETURNING email, nickname, image;`,
		user.Email, user.Nickname, user.Password, user.Image, user.OldNick).Scan(&out.Email, &out.Nickname, &out.Image)
	switch err {
	case sql.ErrNoRows:
		log.Println("Method Update UserData: No rows were returned!")
		exportData := model.UpdateDataExport{
			Email:    user.Email,
			Nickname: user.Nickname,
			Image:    user.Image,
		}
		return exportData, nil
	case nil:
		log.Println("new data of user: ", user)
		return out, nil
	default:
		log.Println("Error updating record:", err)
		return model.UpdateDataExport{}, err
	}
}
func (s *DBService) CheckUser(user model.SigninUser) (out *model.UserRegister, err error) {
	row := s.DB.QueryRow(`SELECT email, nickname, password FROM users WHERE nickname = $1 AND password = $2`, user.Nickname, user.Password)
	switch err := row.Scan(&out.Email, &out.Nickname, &out.Password); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return nil, err
	case nil:
		return out, err		
	default:
		log.Println("Method Signin User: ", err)
		return nil, err
	}
}
