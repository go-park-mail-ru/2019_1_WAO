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

type User struct {
	ID       int    `json:"id, string, omitempty"`
	Email    string `json:"email, omitempty"`
	password string `json:"-"`
	Nick     string `json:"nickname, omitempty"`
	Score    int    `json:"score, string, omitempty"`
	Games    int    `json:"games, string, omitempty"`
	Wins     int    `json:"wins, string, omitempty"`
	Image    string `json:"image, omitempty"`
}

type SessionManager struct {
	// Definition DateBase
	// sessions map[string]*auth.UserData
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		//Initialize DataBase
		// sessions: map[string]*auth.UserData{},
	}
}

func generateToken(in *auth.UserData) (token string, err error) {
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": in.Login,
		"agent":    in.Agent,
		"exp":      time.Now().Add(expires).Unix(),
	})

	tokenString, err := rawToken.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Create JWT for user
func (sm *SessionManager) Create(ctx context.Context, in *auth.UserData) (*auth.Token, error) {
	log.Println("call Create", in)
	token, err := generateToken(in)
	if err != nil {
		log.Println("Token does not create:", err.Error())
		return nil, err
	}
	id := &auth.Token{
		Value: token,
	}
	//Add token to White list of DataBase
	log.Println("Token create: ", id.Value)
	return id, nil

}

// Check validation of token
func (sm *SessionManager) Check(ctx context.Context, tokenString *auth.Token) (*auth.UserData, error) {
	log.Println("call Check", tokenString)

	var err error
	token, err := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, err
		}
		return secret, nil
	})
	if err != nil {
		log.Printf("Unexpected signing method: %v", token.Header["alg"])
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		log.Println("CLAIMS:", claims)
		username, ok := claims["username"]
		if !ok {
			return nil, errors.New("Bad claims: field 'username' not exist")
		}
		// var usernameStr string
		// if usernameStr, ok := username.(string); !ok {
		// 	log.Printf("Error: Cannot convert username (%s) to string", username)
		// }
		user := &auth.UserData{
			Login: username.(string),
		}
		log.Println("Hooooray, Token is exist")
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
