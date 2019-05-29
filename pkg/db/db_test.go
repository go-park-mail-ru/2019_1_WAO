package db

import (
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


func TestGetUsers(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)

	usr, err := conn.GetUsers()
	if err == nil {
		t.Errorf("[%d] wrong connect: got %v, expected %v",
			4, usr, "item.out")
	}
}

func TestGetUser(t *testing.T) {
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

func TestUpdateUser(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	conn := NewDataBase(db)

	out, err := conn.UpdateUser(model.UpdateDataImport{
		Nickname: "testNick",
		Password: "testPass",
		Email: "testEmail",
		Image: "testPath",
		OldNick: "test",
	})
	if err == nil {
		t.Errorf("[%d] wrong connect: got %v, expected %v",
			4, out, "err")
	}
}

func TestCheckUser(t *testing.T) {
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

func TestCreateUser(t *testing.T) {
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