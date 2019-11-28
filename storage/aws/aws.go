package aws

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/Dimitriy14/image-resizing/clients/bucket"
	"github.com/Dimitriy14/image-resizing/config"
	"github.com/Dimitriy14/image-resizing/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func NewStorage(bucketS3 *bucket.S3Client) storage.Storage {
	return &storageImpl{
		bucketS3:         bucketS3,
		bucketName:       config.Conf.AWS.Bucket,
		acl:              config.Conf.AWS.ACL,
		serverEncryption: config.Conf.AWS.ServerSideEncryption,
		awsStorageUrl:    config.Conf.AWS.ImageStorageURL,
	}
}

type storageImpl struct {
	bucketS3         *bucket.S3Client
	bucketName       string
	acl              string
	serverEncryption string
	awsStorageUrl    string
}

// Upload saves a file to aws bucket and returns the url to // the file or an error if there's any
func (s *storageImpl) Upload(filExt string, content []byte) (string, error) {
	fileName := "pictures/" + uuid.New().String() + filExt
	err := s.upload(fileName, content)
	return fmt.Sprintf("%s/%s", s.awsStorageUrl, fileName), err
}

// Download downloads file from aws bucket and returns its content
func (s *storageImpl) Download(addr string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})

	a, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	_, err = s.bucketS3.Downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(a.Path[1:]),
	})

	return buf.Bytes(), err
}

func (s *storageImpl) UploadWithOriginal(filExt string, originalImgContent, resizedImgContent []byte) (string, string, error) {
	var (
		originFileName  = "pictures/" + uuid.New().String() + filExt
		resizedFileName = "pictures/" + uuid.New().String() + filExt
		errc            = make(chan error, 2)
		wg              = new(sync.WaitGroup)
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		errc <- s.upload(originFileName, originalImgContent)
	}()

	go func() {
		defer wg.Done()
		errc <- s.upload(resizedFileName, resizedImgContent)
	}()

	go func() {
		wg.Wait()
		close(errc)
	}()

	for err := range errc {
		if err != nil {
			return "", "", err
		}
	}

	return fmt.Sprintf("%s/%s", s.awsStorageUrl, originFileName), fmt.Sprintf("%s/%s", s.awsStorageUrl, resizedFileName), nil
}

func (s *storageImpl) upload(fileName string, content []byte) error {
	_, err := s.bucketS3.Uploader.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(s.bucketName),
		Key:                  aws.String(fileName),
		ACL:                  aws.String(s.acl),
		Body:                 bytes.NewReader(content),
		ContentType:          aws.String(http.DetectContentType(content)),
		ServerSideEncryption: aws.String(s.serverEncryption),
	})

	return err
}
