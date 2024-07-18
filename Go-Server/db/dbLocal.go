package db

import (
	"context"
	"database/sql"
	"log"
	"websocket/constants"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitLocalMySQL() {
	var err error
	mysqlDBClient, err = sql.Open("mysql", constants.LOCAL_MYSQL_ENDPOINT)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = mysqlDBClient.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
}

func InitLocalDynamoDB() {
	// DynamoDB 로컬 인스턴스 연결 설정
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           constants.LOCAL_DYNAMO_DB_ENDPOINT,
			SigningRegion: constants.BASE_REGION_NAME,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(constants.BASE_REGION_NAME),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
	)
	if err != nil {
		log.Fatalf("Error loading Local AWS DynamoDB config: %v", err)
	}

	// DynamoDB 클라이언트 초기화
	dynamoDBClient = dynamodb.NewFromConfig(cfg)

	// 퀴즈 이미지 데이터 초기화
	InitQuizImageData()
}

func InitLocalS3() {
	// S3 로컬 인스턴스 연결 설정
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           constants.LOCAL_S3_ENDPOINT,
			SigningRegion: constants.BASE_REGION_NAME,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(constants.BASE_REGION_NAME),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(constants.LOCAL_S3_ACCESS_KEY, constants.LOCAL_S3_SECRET_KEY, "dummy")),
	)
	if err != nil {
		log.Fatalf("Error loading Local AWS S3 config: %v", err)
	}

	// S3 클라이언트 초기화
	s3Client = s3.NewFromConfig(cfg)
	s3Bucket = constants.S3_BUCKET
}
