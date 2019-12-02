package bucket

import (
	"github.com/Dimitriy14/image-resizing/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var Client *S3Client

type S3Client struct {
	Uploader   *s3manager.Uploader
	Downloader *s3manager.Downloader
}

func Load() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Conf.AWSRegion),
		Credentials: credentials.NewStaticCredentials(config.Conf.AWSID, config.Conf.AWSSecret, ""),
	})
	if err != nil {
		return err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return err
	}

	Client = &S3Client{
		Uploader:   s3manager.NewUploader(sess),
		Downloader: s3manager.NewDownloader(sess),
	}

	return nil
}
