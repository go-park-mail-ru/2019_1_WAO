package router

import (
	"log"
	"net/http"
	"time"
	"github.com/DmitriyPrischep/backend-WAO/pkg/handlers"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("authMiddleware", r.URL.Path)
		_, err := r.Cookie("session_id")
		if err != nil {
			log.Println("Cookie is not found")
			return
		}
		if _, err := handlers.GetSession(r, Auth); err != nil {
			log.Println("Error checking of session")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("logMiddleware", r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s, %s %s\n",
			r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				log.Println("Recovered error" , err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("CORSMiddleware", r.URL.Path)
		if origin := r.Header.Get("Origin"); origin == frontURL {
			w.Header().Set("Access-Control-Allow-Origin", frontURL)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		next.ServeHTTP(w, r)
	})
}