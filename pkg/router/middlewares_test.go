package router

import (
	"net/http"
	"testing"
)

func TestLogMiddleware(t *testing.T) {
	h1 := LogMiddleware(http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		
	}))
	if h1 == nil {
		t.Errorf("wrong log middleware: got %v", h1)
	} 
	req, _ := http.NewRequest("GET", "https://example.com/foo/", nil)
	h1.ServeHTTP(nil, req)
}

func TestPanicMiddleware(t *testing.T) {
	h1 := PanicMiddleware(http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		panic("Create panic")
	}))
	if h1 == nil {
		t.Errorf("wrong panic middleware: got %v", h1)
	} 
	req, _ := http.NewRequest("GET", "https://example.com/foo/", nil)
	h1.ServeHTTP(nil, req)

	h2 := PanicMiddleware(http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		
	}))
	if h2 == nil {
		t.Errorf("wrong panic middleware: got %v", h1)
	} 
	req, _ = http.NewRequest("GET", "https://example.com/foo/", nil)
	h2.ServeHTTP(nil, req)
}

func TestCORSMiddleware(t *testing.T) {
	frontURL = "https://example.com"
	h1 := CORSMiddleware(http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {

	}))
	if h1 == nil {
		t.Errorf("wrong CORS middleware: got %v", h1)
	} 
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	h1.ServeHTTP(nil, req)

	h2 := CORSMiddleware(http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		
	}))
	if h2 == nil {
		t.Errorf("wrong CORS middleware: got %v", h1)
	} 
	req, _ = http.NewRequest("GET", "https://example.com/foo/", nil)
	h2.ServeHTTP(nil, req)
}