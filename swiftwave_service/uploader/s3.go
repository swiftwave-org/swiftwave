package uploader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/swiftwave-org/swiftwave/local_config"
	"io"
)

func UploadFileToS3(reader io.ReadSeeker, filename, bucket string, config local_config.S3Config) error {
	s3Client, err := GenerateS3Client(config)
	if err != nil {
		return err
	}
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		Body:   reader,
		ACL:    nil,
	})
	return err
}

func DeleteFileFromS3(filename, bucket string, config local_config.S3Config) error {
	s3Client, err := GenerateS3Client(config)
	if err != nil {
		return err
	}
	_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	return err
}

func GenerateS3Client(config local_config.S3Config) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		S3ForcePathStyle: aws.Bool(config.ForcePathStyle),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}
	return s3.New(newSession), nil
}
