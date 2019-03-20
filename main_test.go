package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestCase struct {
	ID         string
	Response   string
	StatusCode int
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("id")
	if key == "42" {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status": 200, "resp": {"user": 42}}`)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `{"status": 500, "err": "db_error"}`)
	}
}

func TestGetUser(t *testing.T) {
	cases := []TestCase{
		TestCase{
			ID:         "42",
			Response:   `{"status": 200, "resp": {"user": 42}}`,
			StatusCode: http.StatusOK,
		},
		TestCase{
			ID:         "500",
			Response:   `{"status": 500, "err": "db_error"}`,
			StatusCode: http.StatusInternalServerError,
		},
	}
	for caseNum, item := range cases {
		url := "http://example.com/api/user?id=" + item.ID
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()

		GetUser(w, req)

		if w.Code != item.StatusCode {
			t.Errorf("[%d] wrong StatusCode: got %d, expected %d",
				caseNum, w.Code, item.StatusCode)
		}

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		bodyStr := string(body)
		if bodyStr != item.Response {
			t.Errorf("[%d] wrong Response: got %+v, expected %+v",
				caseNum, bodyStr, item.Response)
		}
	}
}

type TestCaseUsers struct {
	Response   string
	Token      string
	StatusCode int
}

func TestGetUsers(t *testing.T) {
	cases := []TestCaseUsers{
		TestCaseUsers{
			Response:   `{"id":1,"email":"goshan@pochta.ru","nick":"karlik","scope":119,"games":5,"wins":1,"image":"avatar.jpg"},{"id":2,"email":"pashok@pochta.ru","nick":"joker","scope":200,"games":1,"wins":3,"image":"avatar.jpg"},{"id":3,"email":"karman@pochta.ru","nick":"gopher","scope":88,"games":8,"wins":0,"image":"avatar.jpg"},{"id":4,"email":"support@pochta.ru","nick":"Batman","scope":13,"games":11,"wins":0,"image":"avatar.jpg"}`,
			Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NTM2Nzk2OTksInVzZXJuYW1lIjoia2FybWFuQHBvY2h0YS5ydSJ9.WNj1M95PHOEaowstlU7QR2vnGkKjHVgyiO_gAhE3ZEw",
			StatusCode: http.StatusOK,
		},
		TestCaseUsers{
			Response:   "<a href=\"/v1/login\">See Other</a>.\n\nnull\n",
			Token:      "qwerty.qwerty.qwerty",
			StatusCode: http.StatusSeeOther,
		},
	}
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

		fmt.Println("CODE:", w.Code)

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

		fmt.Printf("\n\n\nEnd test!!!\n\n")
	}
}
