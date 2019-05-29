package driver

import (
	_ "github.com/lib/pq"
	"testing"
)

type TestDriver struct {
	uname string
	pass string
	dbname string
	ssl string
}

func TestConnectSQL(t *testing.T) {
	cases := []TestDriver{
		TestDriver{
			uname: "test",
			pass: "test",
			dbname: "test",
			ssl: "false",
		},
	}
	for caseNum, item := range cases {
		conn, err := ConnectSQL(item.uname, item.pass, item.dbname, item.ssl)
		if conn == nil {
			t.Errorf("[%d] wrong connect: got %v, expected %v",
				caseNum, conn, err)
		}
	}
}