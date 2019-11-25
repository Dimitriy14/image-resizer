package bucket

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Dimitriy14/image-resizing/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

const maxImageSize = 10 << 24 // max image size is 10MB

var Client *S3Client

type S3Client struct {
	S3 *s3.S3
}

func Load() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Conf.AWS.Region),
		Credentials: credentials.NewSharedCredentials("", "Dimidrolio"),
	})
	if err != nil {
		return err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return err
	}

	Client = &S3Client{
		s3: s3.New(sess),
	}

	return nil
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(filename string, content []byte) (string, error) {
	tempFileName := "pictures/" + uuid.New().String() + filepath.Ext(filename)

	bucketS3 := Client.s3

	_, err := bucketS3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("resized-images-yal"),
		Key:    aws.String(tempFileName),
		ACL:    aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:   bytes.NewReader(content),
		//ContentLength:        aws.Int64(int64(len(content))),
		ContentType:          aws.String(http.DetectContentType(content)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}

func handler(w http.ResponseWriter, r *http.Request) {
	maxSize := int64(maxImageSize)

	err := r.ParseMultipartForm(maxSize)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Image too large. Max Size: %v", maxSize)
		return
	}

	file, fileHeader, err := r.FormFile("profile_picture")
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Could not get uploaded file")
		return
	}
	defer file.Close()

	// create an AWS session which can be
	// reused if we're uploading many files

	if err != nil {
		fmt.Fprintf(w, "Could not upload file %s", err)
		return
	}

	fileName, err := UploadFileToS3(file, fileHeader.Filename)
	if err != nil {
		fmt.Fprintf(w, "Could not upload file %s", err)
		return
	}

	fmt.Fprintf(w, "Image uploaded successfully: %v", fileName)
}
