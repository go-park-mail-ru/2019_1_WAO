package db

import (
	"regexp"
	"database/sql"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/DmitriyPrischep/backend-WAO/pkg/methods"
	"github.com/DmitriyPrischep/backend-WAO/pkg/model"
)

type TestAWS struct {
	conn *sql.DB  
	out	methods.UserMethods
}

func TestNewDataBase(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	cases := []TestAWS{
		TestAWS{
			conn: db,
			out: &DBService{DB: db},
		},
	}
	
	for _, item := range cases {
		newDB := NewDataBase(item.conn)
		assert.Equal(t, newDB, item.out)
	}
}

type TestUser struct {
	ID       string   
	Email    string
	Password string
	Nick     string
	Score    int   
	Games    int   
	Wins     int   
	Image    string
}


func TestGetUsers(t *testing.T){
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	cases := []TestUser{
		TestUser{
			ID: "1",
			Nick: "test1",
			Score: 0,
			Games: 0,
			Wins: 0,
		},
		TestUser{
			ID: "2",
			Nick: "test2",
			Score: 1,
			Games: 1,
			Wins: 1,
		},
	}
	mock.ExpectQuery("SELECT id, nickname, score, games, wins FROM users ORDER by score DESC LIMIT 15;").
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "score", "games", "wins"}).
		AddRow(cases[0].ID, cases[0].Nick, cases[0].Score, cases[0].Games, cases[0].Wins).
		AddRow(cases[1].ID, cases[1].Nick, cases[1].Score, cases[1].Games, cases[1].Wins))
	
	testDB := NewDataBase(db)
	out, err := testDB.GetUsers()
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUsersNoRows(t *testing.T){
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := TestUser{
		ID: "1",
		Nick: "test1",
		Score: 0,
		Games: 0,
		Wins: 0,
	}
	mock.ExpectQuery("SELECT id, nickname, score, games, wins FROM users ORDER by score DESC LIMIT 15;").
		WillReturnRows(sqlmock.NewRows([]string{"id", "nickname", "score", "games", "wins"}).
		AddRow(user.ID, user.Nick, user.Score, user.Games, "foo"))
	testDB := NewDataBase(db)
	out, err := testDB.GetUsers()
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUsersEmptyQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)
	usr, err := conn.GetUsers()
	if err == nil {
		t.Errorf(" wrong connect: got %v", usr)
	}
}

func TestGetUserEmptyQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)

	usr, err := conn.GetUser(model.NicknameUser{
		Nickname: "test",
	})
	if usr != nil {
		t.Errorf("[%d] wrong connect: got %v, expected %v",
			4, usr, err)
	}
}

func TestCreateUserEmptyQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)

	out, err := conn.CreateUser(model.UserRegister{
		Email: "testEmail",
		Nickname: "testNick",
		Password: "testPass",
	})
	if err == nil {
		t.Errorf("wrong connect: got %v, expected %v", out, "err")
	}
}

func TestCreateUser(t *testing.T){
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UserRegister{
		Email: "test@test.ru",
		Nickname: "test1",
		Password: "test_pass",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (email, nickname, password, score, games, wins, image) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING nickname;`)).
		WithArgs(user.Email, user.Nickname, user.Password, 0, 0, 0, "default_image.png").
		WillReturnRows(sqlmock.NewRows([]string{"nickname"}).
		AddRow(user.Nickname))
	testDB := NewDataBase(db)
	out, err := testDB.CreateUser(user)
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := TestUser{
		ID: "1",
		Email: "test@test.ru",
		Nick: "test1",
		Score: 0,
		Games: 0,
		Wins: 0,
		Image: "avatar.jpg",
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, nickname, score, games, wins, image FROM users WHERE nickname = $1;")).
		WithArgs(user.Nick).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "nickname", "score", "games", "wins", "image"}).
		AddRow(user.ID, user.Email, user.Nick, user.Score, user.Games, user.Wins, user.Image))
	testDB := NewDataBase(db)
	out, err := testDB.GetUser(model.NicknameUser{Nickname: user.Nick})
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, nickname, score, games, wins, image FROM users WHERE nickname = $1;")).
		WithArgs("foo").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "nickname", "score", "games", "wins", "image"}))
	testDB := NewDataBase(db)
	out, err := testDB.GetUser(model.NicknameUser{Nickname: "foo"})
	if err == nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}


func TestCheckUserEmptyQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)

	out, err := conn.CheckUser(model.SigninUser{
		Nickname: "testNick",
		Password: "testPass",
	})
	if err == nil {
		t.Errorf("wrong connect: got %v, expected %v", out, "err")
	}
}

func TestCheckUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UserRegister{
		Email: "test@test.ru",
		Nickname: "test",
		Password: "test_pass",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT email, nickname, password FROM users WHERE nickname = $1`)).
		WithArgs(user.Nickname).
		WillReturnRows(sqlmock.NewRows([]string{"email", "nickname", "password"}).
		AddRow(user.Email, user.Nickname, user.Password))
	testDB := NewDataBase(db)
	out, err := testDB.CheckUser(model.SigninUser{
		Nickname: user.Nickname,
		Password: user.Password,
	})
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckUserNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UserRegister{
		Email: "test@test.ru",
		Nickname: "test",
		Password: "test_pass",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT email, nickname, password FROM users WHERE nickname = $1;`)).
		WithArgs("Foo").
		WillReturnRows(sqlmock.NewRows([]string{"email", "nickname", "password"}))
	testDB := NewDataBase(db)
	out, err := testDB.CheckUser(model.SigninUser{
		Nickname: "Foo",
		Password: user.Password,
	})
	if err == nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UpdateDataImport{
		Email: "test@test.ru",
		Nickname: "test",
		Password: "test_pass",
		Image: "avatar.jpg",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
	UPDATE users SET
		email = COALESCE(NULLIF($1, ''), email),
		password = COALESCE(NULLIF($3, ''), password),
		image = COALESCE(NULLIF($4, ''), image)
	WHERE nickname = $2
	AND  (NULLIF($1, '') IS NOT NULL AND NULLIF($1, '') IS DISTINCT FROM email OR
		NULLIF($3, '') IS NOT NULL AND NULLIF($3, '') IS DISTINCT FROM password OR
		NULLIF($4, '') IS NOT NULL AND NULLIF($4, '') IS DISTINCT FROM image)
	RETURNING email, nickname, image;`)).
		WithArgs(user.Email, user.Nickname, user.Password, user.Image).
		WillReturnRows(sqlmock.NewRows([]string{"email", "nickname", "image"}).
		AddRow(user.Email, user.Nickname, user.Image))
	testDB := NewDataBase(db)
	out, err := testDB.UpdateUser(user)
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateUserNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UpdateDataImport{
		Email: "test@test.ru",
		Nickname: "test",
		Password: "test_pass",
		Image: "avatar.jpg",
	}
	mock.ExpectQuery(regexp.QuoteMeta(`
	UPDATE users SET
		email = COALESCE(NULLIF($1, ''), email),
		password = COALESCE(NULLIF($3, ''), password),
		image = COALESCE(NULLIF($4, ''), image)
	WHERE nickname = $2
	AND  (NULLIF($1, '') IS NOT NULL AND NULLIF($1, '') IS DISTINCT FROM email OR
		NULLIF($3, '') IS NOT NULL AND NULLIF($3, '') IS DISTINCT FROM password OR
		NULLIF($4, '') IS NOT NULL AND NULLIF($4, '') IS DISTINCT FROM image)
	RETURNING email, nickname, image;`)).
		WithArgs(user.Email, user.Nickname, user.Password, user.Image).
		WillReturnRows(sqlmock.NewRows([]string{"email", "nickname", "image"}))
	testDB := NewDataBase(db)
	out, err := testDB.UpdateUser(user)
	if err != nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateUserEmptyQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	user := model.UpdateDataImport{
		Email: "test@test.ru",
		Nickname: "test",
		Password: "test_pass",
		Image: "avatar.jpg",
	}
	testDB := NewDataBase(db)
	out, err := testDB.UpdateUser(user)
	if err == nil {
		t.Error(out, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}