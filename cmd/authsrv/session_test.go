package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"golang.org/x/net/context"
)

type TestAWS struct {
	out	*SessionManager
}

func TestNewSessionManager(t *testing.T) {
	cases := []TestAWS{
		TestAWS{
			out: &SessionManager{},
		},
	}
	
	for _, item := range cases {
		sess := NewSessionManager()
		assert.Equal(t, sess, item.out)
	}
}

type TestSession struct {
	sess *auth.UserData
	token string
}

func TestGenerateToken(t *testing.T) {
	cases := []TestSession{
		TestSession{
			sess: &auth.UserData{
				Login: "testLogin",
				Password: "testPass",
				Agent: "testAgent",
			},
			token: "Error",
		},
	}
	for caseNum, item := range cases {
		token, err := generateToken(item.sess)

		if err != nil {
			t.Errorf("[%d] wrong token: got %v, expected %v",
				caseNum, token, item.token)
		}
	}
}

type TestCreateSession struct {
	ctx context.Context
	in *auth.UserData
	out *auth.Token
}

func TestCreateToken(t *testing.T) {
	cases := []TestCreateSession{
		TestCreateSession{
			ctx: context.Background(),
			in: &auth.UserData{
				Login: "testNick",
				Password: "testPass",
				Agent: "testAgent",
			},
			out: &auth.Token{
				Value: "token",
			},
		},
	}
	sess := NewSessionManager()
	for caseNum, item := range cases {
		out, err := sess.Create(item.ctx, item.in)
		if err != nil {
			t.Errorf("[%d] wrong create token: got %v, expected %v",
				caseNum, out, item.out)
		}
	}
}

type TestDeleteSession struct {
	ctx context.Context
	in *auth.Token
	out *auth.Nothing
}

func TestDeleteToken(t *testing.T) {
	cases := []TestDeleteSession{
		TestDeleteSession{
			ctx: context.Background(),
			in: &auth.Token{
				Value: "AuthToken",
			},
			out: &auth.Nothing{
				Null: true,
			},
		},
	}
	sess := NewSessionManager()
	for caseNum, item := range cases {
		out, err := sess.Delete(item.ctx, item.in)
		if err != nil {
			t.Errorf("[%d] wrong delete token: got %v, expected %v",
				caseNum, out, item.out)
		}
	}
}

type TestCheckSession struct {
	ctx context.Context
	in *auth.Token
	out *auth.UserData
}

func TestCheckToken(t *testing.T) {
	cases := []TestCheckSession{
		TestCheckSession{
			ctx: context.Background(),
			out: &auth.UserData{
				Login: "testNick",
				Password: "testPass",
				Agent: "testAgent",
			},
			in: &auth.Token{
				Value: "token",
			},
		},
	}
	sess := NewSessionManager()
	for caseNum, item := range cases {
		token, err := sess.Create(item.ctx, item.out)

		res, err := sess.Check(item.ctx, token)
		if err != nil {
			t.Errorf("[%d] wrong check token: got %v, expected %v",
				caseNum, token, res)
		}
	}
}