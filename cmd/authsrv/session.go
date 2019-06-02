package main

import (
	"errors"
	"log"
	"time"

	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	expiration = 360 * time.Minute
)

type SessionManager struct {
	// Definition DateBase
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		//Initialize DataBase
	}
}

func generateToken(in *auth.UserData) (token string, err error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": in.Login,
		"id":       in.Id,
		"agent":    in.Agent,
		"exp":      time.Now().Add(expiration).Unix(),
	})
	tokenString, err := rawToken.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Create JWT for user
func (sm *SessionManager) Create(ctx context.Context, in *auth.UserData) (*auth.Token, error) {
	log.Println("call Create")
	token, err := generateToken(in)
	if err != nil {
		log.Println("Token does not create:", err)
		return nil, err
	}
	id := &auth.Token{
		Value: token,
	}
	//Add token to White list of DataBase
	return id, nil

}

// Check validation of token
func (sm *SessionManager) Check(ctx context.Context, in *auth.Token) (*auth.UserData, error) {
	log.Println("call Check", in)
	var err error
	token, err := jwt.Parse(in.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, err
		}
		return []byte(secret), nil
	})
	if err != nil {
		log.Printf("Unexpected signing method: %v", token.Header["alg"])
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"]
		if !ok {
			return nil, errors.New("Bad claims: field 'username' not exist")
		}
		id, ok := claims["id"]
		if !ok {
			return nil, errors.New("Bad claims: field 'username' not exist")
		}

		user := &auth.UserData{
			Login: username.(string),
			Id:    id.(string),
		}
		return user, nil
	}
	return nil, grpc.Errorf(codes.NotFound, "session not found")
}

// Delete token
func (sm *SessionManager) Delete(ctx context.Context, in *auth.Token) (*auth.Nothing, error) {
	log.Println("call Delete", in)
	//Delete from WhiteList of DataBase
	return &auth.Nothing{Null: true}, nil
}
