package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

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
			Email:    "karman@pochta.ru",
			password: "12345",
			Nick:     "gopher",
			Scope:    88,
			Games:    8,
			Wins:     0,
			Image:    "avatar.jpg",
		},
		User{
			ID:       4,
			Email:    "support@pochta.ru",
			password: "12345",
			Nick:     "Batman",
			Scope:    13,
			Games:    11,
			Wins:     0,
			Image:    "avatar.jpg",
		},
	)
}

type TestCaseUsers struct {
	Response   string
	Token      string
	StatusCode int
}

func TestGetUsers(t *testing.T) {
	cases := []TestCaseUsers{
		TestCaseUsers{
			Response:   `[{"id":1,"email":"goshan@pochta.ru","nick":"karlik","scope":119,"games":5,"wins":1,"image":"avatar.jpg"},{"id":2,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":1,"wins":3,"image":"avatar.jpg"},{"id":3,"email":"karman@pochta.ru","nick":"gopher","scope":88,"games":8,"wins":0,"image":"avatar.jpg"},{"id":4,"email":"support@pochta.ru","nick":"Batman","scope":13,"games":11,"wins":0,"image":"avatar.jpg"}]` + "\n",
			Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NTM2Nzk2OTksInVzZXJuYW1lIjoia2FybWFuQHBvY2h0YS5ydSJ9.WNj1M95PHOEaowstlU7QR2vnGkKjHVgyiO_gAhE3ZEw",
			StatusCode: http.StatusOK,
		},
		TestCaseUsers{
			Response:   "<a href=\"/v1/login\">See Other</a>.\n\n",
			Token:      "qwerty.qwerty.qwerty",
			StatusCode: http.StatusSeeOther,
		},
	}
	mockUsers()
	for caseNum, item := range cases {
		url := "http://test.go/api/users"
		req := httptest.NewRequest("GET", url, nil)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    item.Token,
			HttpOnly: true,
		}
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		GetUsers(w, req)

		if w.Code != item.StatusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.StatusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body[:])
		if bodyStr != item.Response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.Response)
		}
	}
	Users = nil
}

type TestCaseUser struct {
	ID         string
	Response   string
	Token      string
	StatusCode int
}

func TestGetUser(t *testing.T) {
	cases := []TestCaseUser{
		TestCaseUser{
			ID:         "2",
			Response:   "Player profile joker<br><img src=\"/data/2/avatar.jpg\"/>",
			Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NTM2Nzk2OTksInVzZXJuYW1lIjoia2FybWFuQHBvY2h0YS5ydSJ9.WNj1M95PHOEaowstlU7QR2vnGkKjHVgyiO_gAhE3ZEw",
			StatusCode: http.StatusOK,
		},
		TestCaseUser{
			ID:         "1",
			Response:   "<a href=\"/v1/login\">See Other</a>.\n\n",
			Token:      "qwerty.qwerty.qwerty",
			StatusCode: http.StatusSeeOther,
		},
	}

	mockUsers()
	for caseNum, item := range cases {
		url := "http://test.go/api/users/id-" + item.ID
		req := httptest.NewRequest("GET", url, nil)
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    item.Token,
			HttpOnly: true,
		}
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

		//Hack to try to fake gorilla/mux vars
		vars := map[string]string{
			"id": item.ID,
		}

		// CHANGE THIS LINE!!!
		req = mux.SetURLVars(req, vars)

		GetUser(w, req)

		if w.Code != item.StatusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.StatusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body[:])
		if bodyStr != item.Response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.Response)
		}
	}
	Users = nil
}
