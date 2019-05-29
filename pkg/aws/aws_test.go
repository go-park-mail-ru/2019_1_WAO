package aws

import (
	"testing"
	"net/textproto"
	"net/http"
)

type TestAWS struct {
	conn   *ConnectSetting
	out	string
}


func TestNewConnectAWS(t *testing.T) {
	cases := []TestAWS{
		TestAWS{
			conn: &ConnectSetting{
				AccessKeyID: "",
				SecretAccessKey: "",
				Token: "",
				Region: "",
				NameBucket: "",
				PathRootDir: "",
			},
			out: "nothing",
		},
	}
	for caseNum, item := range cases {
		aws := NewConnectAWS(item.conn)
		if aws == nil {
			t.Errorf("[%d] wrong connect: got %v, expected %v",
				caseNum, aws, item.out)
		}
	}
}

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader
	Size     int64 // Go 1.9
	// contains filtered or unexported fields
}

type Form struct {
	Value map[string][]string
	File  map[string][]*FileHeader
}


func TestUploadImage(t *testing.T) {
	set := &ConnectSetting{
		AccessKeyID: "",
		SecretAccessKey: "",
		Token: "",
		Region: "",
		NameBucket: "",
		PathRootDir: "",
	}
	conn := NewConnectAWS(set)

	r := http.Request{
	}


	ffr, _, err := r.FormFile("image"); 
	_, err = conn.UploadImage(ffr, "img", 0)
	if err == nil{
		t.Errorf("wrong upload: got %v, expected %v",
				err, nil)
	}

}