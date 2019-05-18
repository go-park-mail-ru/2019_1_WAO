package aws

import (
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

func NewConnectAWS(keyID, secret, token, region, name, pathRoot string) *ConnectSetting {
	return &ConnectSetting{
		AccessKeyID: keyID,
		SecretAccessKey: secret,
		Token: token,
		Region: region,
		NameBucket: name,
		PathRootDir: pathRoot,
	}
}

func (s *ConnectSetting) UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (url string, err error) {
	str := fileHeader.Filename + string(time.Now().Format("15:04:05.00000"))
	hash := sha1.New()
	hash.Write([]byte(str))
	hashSHA1 := hex.EncodeToString(hash.Sum(nil))
	log.Println("HASH:", hashSHA1)

	creds := credentials.NewStaticCredentials(s.AccessKeyID , s.SecretAccessKey, s.Token)
	_, err = creds.Get()
	if err != nil {
		log.Println("Error credentials: ", err)
	}

	cfg := aws.NewConfig().WithRegion("us-east-2").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	path := "/media/" + hashSHA1
	params := &s3.PutObjectInput{
		Bucket:        aws.String("waojump"),
		Key:           aws.String(path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	// paramsAcl := &s3.PutObjectAcl{
	// 	ACL:       String("public-read"),
	// 	Key:       String("any-file"),
	// 	VersionId: String("2"),
	// 	Bucket:    String("other-bucket"),
	// }

	_, err = svc.PutObject(params)
	if err != nil {
		log.Println("Put object error: ", err.Error())
	}
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("waojump"),
		Key:    aws.String("media/" + hashSHA1),
	})
	url, err = req.Presign(15 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}
	return 
}