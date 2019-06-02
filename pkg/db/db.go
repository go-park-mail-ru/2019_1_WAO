package db

import (
	"database/sql"
	"log"
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

func (s *DBService) GetUsers() (users []model.Player, err error) {
	rows, err := s.DB.Query("SELECT id, nickname, score, games, wins FROM users ORDER by score DESC LIMIT 15;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		user := model.Player{}
		err := rows.Scan(&user.ID, &user.Nick, &user.Score, &user.Games, &user.Wins)
		if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, user)
	}
	return users, err
}
func (s *DBService) GetUser(userdata model.NicknameUser) (user *model.User, err error) {
	tmp := model.User{}
	row := s.DB.QueryRow(`SELECT id, email, nickname, score, games, wins, image 
						FROM users WHERE nickname = $1;`, userdata.Nickname)

	switch err := row.Scan(&tmp.ID, &tmp.Email, &tmp.Nick, &tmp.Score,
		&tmp.Games, &tmp.Wins, &tmp.Image); err {
	case sql.ErrNoRows:
		log.Println("Method GetUser: No rows were returned!")
		return nil, err
	case nil:
		user = &tmp
		return user, err
	default:
		return nil, err
	}
}

func (s *DBService) CreateUser(user model.UserRegister) (out *model.RegUser, err error) {
	tmp := model.RegUser{}
	err = s.DB.QueryRow(`INSERT INTO users (email, nickname, password, score, games, wins, image)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, nickname;`,
	user.Email, user.Nickname, user.Password, 0, 0, 0, "default_image.png").Scan(&tmp.ID, &tmp.Nickname)
	if err != nil {
		return nil, err
	}
	return &tmp, nil
}

func (s *DBService) UpdateUser(user model.UpdateDataImport) (out model.UpdateDataExport, err error) {
	err = s.DB.QueryRow(`
	UPDATE users SET
		email = COALESCE(NULLIF($1, ''), email),
		password = COALESCE(NULLIF($3, ''), password),
		image = COALESCE(NULLIF($4, ''), image)
	WHERE nickname = $2
	AND  (NULLIF($1, '') IS NOT NULL AND NULLIF($1, '') IS DISTINCT FROM email OR
		NULLIF($3, '') IS NOT NULL AND NULLIF($3, '') IS DISTINCT FROM password OR
		NULLIF($4, '') IS NOT NULL AND NULLIF($4, '') IS DISTINCT FROM image)
	RETURNING email, nickname, image;`,
		user.Email, user.Nickname, user.Password, user.Image).Scan(&out.Email, &out.Nickname, &out.Image)
	switch err {
	case sql.ErrNoRows:
		exportData := model.UpdateDataExport{
			Email:    user.Email,
			Nickname: user.Nickname,
			Image:    user.Image,
		}
		log.Println("DB: no rows return")
		return exportData, nil
	case nil:
		log.Println("DB: new data of user: ", user)
		return out, nil
	default:
		return model.UpdateDataExport{}, err
	}
}
func (s *DBService) CheckUser(user model.SigninUser) (out *model.UserRegister, err error) {
	tmp := model.UserRegister{}
	row := s.DB.QueryRow(`SELECT id, email, nickname, password FROM users WHERE nickname = $1;`, user.Nickname)
	switch err := row.Scan(&tmp.ID, &tmp.Email, &tmp.Nickname, &tmp.Password); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return nil, err
	case nil:
		out = &tmp
		return out, err		
	default:
		return nil, err
	}
}
