package router

import (
	"net/http"
	"testing"
	"log"
	"github.com/DmitriyPrischep/backend-WAO/pkg/driver"
	"github.com/DmitriyPrischep/backend-WAO/pkg/auth"
	"github.com/DmitriyPrischep/backend-WAO/pkg/aws"
	"github.com/DATA-DOG/go-sqlmock"
)

type TestRouter struct {
	prefix   string
	urlCORS      string
	urlImage string
	serviceSession auth.AuthCheckerClient
	db *driver.DB
	setting *aws.ConnectSetting
	handler http.Handler
}

func TestCreateRouter(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	cases := []TestRouter{
		TestRouter{
			prefix: "/api",
			urlCORS: "localhost",
			db: &driver.DB{
				DB: db,
			},
			urlImage: "http://default",
			serviceSession: auth.NewAuthCheckerClient(nil),
			setting: &aws.ConnectSetting{
				AccessKeyID:     "key",
				SecretAccessKey: "Access",
				Token:           "",
				Region:          "",
				NameBucket:      "",
				PathRootDir:     "",
			},
			handler: http.NewServeMux(),
		},
	}
	for caseNum, item := range cases {
		hand := CreateRouter(item.prefix, item.urlCORS, item.urlImage, item.serviceSession, item.db, item.setting)

		log.Printf("Type router: %T\n%v", hand, hand)
		if hand == nil {
			t.Errorf("[%d] wrong router: got %v, expected %v",
				caseNum, hand, "http.Handler")
		}
	}
}