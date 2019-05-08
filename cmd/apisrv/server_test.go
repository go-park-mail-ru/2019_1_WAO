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

type TestCaseUsers struct {
	response   string
	login      string
	statusCode int
}

func TestGetUsers(t *testing.T) {
	cases := []TestCaseUsers{
		TestCaseUsers{
			response:   `[{"id":1,"email":"goshan@pochta.ru","nick":"karlik","scope":119,"games":5,"wins":1,"image":""},{"id":3,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":11,"wins":3,"image":"/data/3/avatar.jpg"}]` + "\n",
			statusCode: http.StatusOK,
		},
		TestCaseUsers{
			response:   `[{"id":1,"email":"goshan@pochta.ru","nick":"karlik","scope":119,"games":5,"wins":1,"image":""},{"id":3,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":11,"wins":3,"image":"/data/3/avatar.jpg"}]` + "\n",
			login:      "__invalid_login__",
			statusCode: http.StatusOK,
		},
	}

	for caseNum, item := range cases {
		url := "http://test.go/api/users"
		req := httptest.NewRequest("GET", url, nil)
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
}

type TestCaseUser struct {
	nickname   string
	password   string
	response   string
	statusCode int
}

func TestGetUser(t *testing.T) {
	cases := []TestCaseUser{
		TestCaseUser{
			nickname:   "joker",
			password:   "12345",
			response:   `{"id":3,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":11,"wins":3,"image":"/data/3/avatar.jpg"}` + "\n",
			statusCode: http.StatusOK,
		},
		TestCaseUser{
			nickname:   "___undefined_nick",
			password:   "qwerty",
			response:   `{"error": "This user is not found"}` + "\n",
			statusCode: http.StatusNotFound,
		},
	}

	for caseNum, item := range cases {
		url := "http://test.go/api/users/id-" + item.nickname
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		//Hack to try to fake gorilla/mux vars
		vars := map[string]string{
			"login":    item.nickname,
			"password": item.password,
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
			statusCode: http.StatusBadRequest,
			player: User{
				Email:    "qwerty@test.go",
				password: "12345",
				Nick:     "mynick",
			},
		},
		NewUser{
			response:   "",
			statusCode: http.StatusBadRequest,
			player: User{
				Email:    "",
				password: "",
				Nick:     "",
			},
		},
	}

	for caseNum, item := range cases {
		urlString := "http://test.go/api/users/"

		form := url.Values{}
		form.Add("email", item.player.Email)
		form.Add("password", item.player.password)
		form.Add("nickname", item.player.Nick)

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

	for caseNum, _ := range cases {
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

		// if status, _ := checkAuthorization(*req); status == item.status {
		// 	t.Errorf("[%d] wrong StatusCode: got %v, expected %v",
		// 		caseNum, status, item.status)
		// }
	}
}

type OldUser struct {
	response   string
	statusCode int
	player     User
}

func TestdeleteUser(t *testing.T) {
	cases := []OldUser{
		OldUser{
			response:   "",
			statusCode: http.StatusBadRequest,
			player: User{
				Nick: "joker",
			},
		},
		OldUser{
			response:   "",
			statusCode: http.StatusNotImplemented,
			player: User{
				Nick: "notor",
			},
		},
	}

	for caseNum, item := range cases {
		url := "http://test.go/api/users/id-" + item.player.Nick

		req := httptest.NewRequest("DELETE", url, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		deleteUser(w, req)

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
}

func TestupdateUser(t *testing.T) {
	cases := []OldUser{
		OldUser{
			response:   "",
			statusCode: http.StatusBadRequest,
			player: User{
				Nick: "joker",
			},
		},
	}

	for caseNum, item := range cases {
		url := "http://test.go/api/users/id-" + item.player.Nick

		req := httptest.NewRequest("DELETE", url, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		updateUser(w, req)

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
}

type Session struct {
	response   string
	statusCode int
	session    string
	nickname   string
}

func TestcheckSession(t *testing.T) {
	cases := []Session{
		Session{
			response:   "",
			statusCode: http.StatusBadRequest,
			nickname:   "joker",
			session:    "token",
		},
	}

	for caseNum, item := range cases {
		url := "http://test.go/api/users/id-" + item.player.Nick

		req := httptest.NewRequest("DELETE", url, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		updateUser(w, req)

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
}
