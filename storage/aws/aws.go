package aws

import (
	"bytes"
	"net/http"
	"path/filepath"

	"github.com/Dimitriy14/image-resizing/config"

	"github.com/Dimitriy14/image-resizing/clients/bucket"
	"github.com/Dimitriy14/image-resizing/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func NewUploader(bucketS3 bucket.S3Client) storage.Uploader {
	return &uploaderImpl{bucketS3: bucketS3}
}

type uploaderImpl struct {
	bucketS3 bucket.S3Client
}

// Upload saves a file to aws bucket and returns the url to // the file or an error if there's any
func (u *uploaderImpl) Upload(filename string, content []byte) (string, error) {
	tempFileName := "pictures/" + uuid.New().String() + filepath.Ext(filename)

	_, err := u.bucketS3.S3.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(config.Conf.AWS.Bucket),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String(config.Conf.AWS.ACL),
		Body:                 bytes.NewReader(content),
		ContentType:          aws.String(http.DetectContentType(content)),
		ServerSideEncryption: aws.String(config.Conf.AWS.ServerSideEncryption),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}
