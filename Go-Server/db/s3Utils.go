package db

import (
	"context"
	"fmt"
	"time"
	"websocket/constants"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client *s3.Client
	s3Bucket string
)

func InitS3() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	// S3 클라이언트 초기화
	s3Client = s3.NewFromConfig(cfg)
	s3Bucket = constants.S3_BUCKET
}

func GetPresignedURLForImage(key string) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	request, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = constants.PRESIGNED_URL_EXPIRATION_MINUTE * time.Minute // Presigned URL expiration time
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return request.URL, nil
}
