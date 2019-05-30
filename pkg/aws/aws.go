package aws

import (
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"time"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ConnectSetting struct {
	AccessKeyID string
	SecretAccessKey string
	Token string
	Region string
	NameBucket string
	PathRootDir string
}

func NewConnectAWS(conn *ConnectSetting) *ConnectSetting {
	return &ConnectSetting{
		AccessKeyID: conn.AccessKeyID,
		SecretAccessKey: conn.SecretAccessKey,
		Token: conn.Token,
		Region: conn.Region,
		NameBucket: conn.NameBucket,
		PathRootDir: conn.PathRootDir,
	}
}

func (s *ConnectSetting) UploadImage(file multipart.File, filename string, size int64) (url string, err error) {
	str := filename + string(time.Now().Format("15:04:05"))
	hash := sha1.New()
	hash.Write([]byte(str))
	hashSHA1 := hex.EncodeToString(hash.Sum(nil))
	creds := credentials.NewStaticCredentials(s.AccessKeyID , s.SecretAccessKey, s.Token)
	_, err = creds.Get()
	if err != nil {
		log.Println("Error credentials: ", err)
	}

	cfg := aws.NewConfig().WithRegion(s.Region).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	buffer := make([]byte, size)
	if size == 0{
		log.Println("File is empty")
		return "", errors.New("File is empty")
	}
	file.Read(buffer)
	fileType := http.DetectContentType(buffer)

	path := "/" + s.PathRootDir + hashSHA1
	params := &s3.PutObjectInput{
		Bucket:        aws.String(s.NameBucket),
		Key:           aws.String(path),
		ACL:           aws.String("public-read"),
		ContentType:   aws.String(fileType),
		Body:          bytes.NewReader(buffer),
		Metadata: map[string]*string{
			"key-f": aws.String("value-bar"),
		},
	}

	_, err = svc.PutObject(params)
	if err != nil {
		log.Println("Put object error: ", err.Error())
		return
	}
	url = hashSHA1
	return 
}