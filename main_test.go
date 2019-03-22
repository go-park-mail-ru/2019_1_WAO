package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func mockUsers() {
	Users = append(Users,
		User{
			ID:       1,
			Email:    "goshan@pochta.ru",
			password: "12345",
			Nick:     "karlik",
			Scope:    119,
			Games:    5,
			Wins:     1,
			Image:    "avatar.jpg",
		},
		User{
			ID:       2,
			Email:    "pashok@pochta.ru",
			password: "12345",
			Nick:     "joker",
			Scope:    200,
			Games:    1,
			Wins:     3,
			Image:    "avatar.jpg",
		},
		User{
			ID:       3,
			Email:    "test_login",
			password: "12345",
			Nick:     "Watermar",
			Scope:    1,
			Games:    1,
			Wins:     1,
			Image:    "avatar.jpg",
		},
	)
}

type TestCaseUsers struct {
	response   string
	login      string
	statusCode int
}

func TestGetUsers(t *testing.T) {
	cases := []TestCaseUsers{
		TestCaseUsers{
			response:   `[{"id":1,"email":"goshan@pochta.ru","nick":"karlik","scope":119,"games":5,"wins":1,"image":"avatar.jpg"},{"id":2,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":1,"wins":3,"image":"avatar.jpg"},{"id":3,"email":"test_login","nick":"Watermar","scope":1,"games":1,"wins":1,"image":"avatar.jpg"}]` + "\n",
			login:      "test_login",
			statusCode: http.StatusOK,
		},
		TestCaseUsers{
			response:   "<a href=\"/v1/login\">See Other</a>.\n\n",
			login:      "__invalid_login__",
			statusCode: http.StatusSeeOther,
		},
	}
	mockUsers()

	for caseNum, item := range cases {
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": item.login,
			"exp":      time.Now().Add(1 * time.Minute).Unix(),
		})
		secret = nil
		secret = []byte("test_secret")

		token, err := rawToken.SignedString(secret)
		if caseNum == 1 {
			token, err = rawToken.SignedString([]byte("fooo"))
		}
		if err != nil {
			log.Println([]byte("Error: Token was not create!" + err.Error()))
		}

		url := "http://test.go/api/users"
		req := httptest.NewRequest("GET", url, nil)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    token,
			HttpOnly: true,
		}
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		GetUsers(w, req)

		if w.Code != item.statusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.statusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body[:])
		if bodyStr != item.response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.response)
		}
	}
	Users = nil
}

type TestCaseUser struct {
	id         string
	email      string
	response   string
	statusCode int
}

func TestGetUser(t *testing.T) {
	cases := []TestCaseUser{
		TestCaseUser{
			id:         "2",
			email:      "pashok@pochta.ru",
			response:   "Player profile joker<br><img src=\"/data/2/avatar.jpg\"/>",
			statusCode: http.StatusOK,
		},
		TestCaseUser{
			id:         "1",
			email:      "goshan@pochta.ru",
			response:   "<a href=\"/v1/login\">See Other</a>.\n\n",
			statusCode: http.StatusSeeOther,
		},
	}

	mockUsers()
	for caseNum, item := range cases {
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": item.email,
			"exp":      time.Now().Add(1 * time.Minute).Unix(),
		})
		secret = nil
		secret = []byte("test_secret")

		token, err := rawToken.SignedString(secret)
		if caseNum == 1 {
			token, err = rawToken.SignedString([]byte("fooo"))
		}
		if err != nil {
			log.Println([]byte("Error: Token was not create!" + err.Error()))
		}

		url := "http://test.go/api/users/id-" + item.id
		req := httptest.NewRequest("GET", url, nil)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    token,
			HttpOnly: true,
		}
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		//Hack to try to fake gorilla/mux vars
		vars := map[string]string{
			"id": item.id,
		}
		req = mux.SetURLVars(req, vars)

		GetUser(w, req)
		if w.Code != item.statusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.statusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body[:])
		if bodyStr != item.response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.response)
		}
	}
	Users = nil
}

type NewUser struct {
	response   string
	statusCode int
	player     User
}

func TestCreateUser(t *testing.T) {
	cases := []NewUser{
		NewUser{
			response:   "",
			statusCode: http.StatusOK,
			player: User{
				Email:    "qwerty@test.go",
				password: "12345",
			},
		},
		NewUser{
			response:   `{"error": "Invalid data"}`,
			statusCode: http.StatusOK,
			player: User{
				Email:    "",
				password: "",
			},
		},
	}

	mockUsers()
	for caseNum, item := range cases {
		urlString := "http://test.go/api/users/"

		form := url.Values{}
		form.Add("email", item.player.Email)
		form.Add("password", item.player.password)

		req := httptest.NewRequest("POST", urlString, nil)
		req.Form = form
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		CreateUser(w, req)

		if w.Code != item.statusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.statusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body[:])
		if bodyStr != item.response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.response)
		}
	}
	Users = nil
}

type auth struct {
	status bool
}

func TestcheckAuthorization(t *testing.T) {
	cases := []auth{
		auth{
			status: true,
		},
		auth{
			status: false,
		},
	}

	for caseNum, item := range cases {
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": "test_login",
			"exp":      time.Now().Add(1 * time.Minute).Unix(),
		})
		secret = nil
		secret = []byte("test_secret")

		token, err := rawToken.SignedString(secret)
		if caseNum == 1 {
			token, err = rawToken.SignedString([]byte("fooo"))
		}
		if err != nil {
			log.Println([]byte("Error: Token was not create!" + err.Error()))
		}

		url := "http://test.go/api/users"
		req := httptest.NewRequest("GET", url, nil)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    token,
			HttpOnly: true,
		}
		req.AddCookie(cookie)

		if status, _ := checkAuthorization(*req); status == item.status {
			t.Errorf("[%d] wrong StatusCode: got %v, expected %v",
				caseNum, status, item.status)
		}
	}
}
