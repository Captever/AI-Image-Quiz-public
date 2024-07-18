package db

import (
	"context"
	"fmt"
	"log"
	"websocket/constants"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var dynamoDBClient *dynamodb.Client

func InitDynamoDB() {
	// Initialize DynamoDB client
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(constants.BASE_REGION_NAME))
	if err != nil {
		log.Fatalf("Error loading AWS config: %v", err)
	}

	// DynamoDB 클라이언트 초기화
	dynamoDBClient = dynamodb.NewFromConfig(cfg)

	// 퀴즈 이미지 데이터 초기화
	InitQuizImageData()
}

func CreateSession(userId, sessionId string) error {
	_, err := dynamoDBClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(constants.DYNAMO_USER_SESSION_TABLE_NAME),
		Item: map[string]types.AttributeValue{
			"session_id": &types.AttributeValueMemberS{Value: sessionId},
			"user_id":    &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func DeleteSession(sessionId string) error {
	_, err := dynamoDBClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(constants.DYNAMO_USER_SESSION_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"session_id": &types.AttributeValueMemberS{Value: sessionId},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func UpdateUserScore(clientUUID, scoreAmount string) error {
	// TODO: DynamoDB에 값이 반영되도록
	// - user_id가 존재하지 않는다면 새 record 생성
	// - user_id가 존재한다면 기존 record의 값 변경
	_, err := dynamoDBClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(constants.DYNAMO_USER_SCORE_TABLE_NAME),
		Item: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: clientUUID},
			"score":   &types.AttributeValueMemberN{Value: scoreAmount},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user score: %w", err)
	}
	return nil
}
