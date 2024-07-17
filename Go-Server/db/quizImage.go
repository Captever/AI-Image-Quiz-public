package db

import (
	"context"
	"fmt"
	"math/rand"
	"path"
	"sort"
	"strconv"
	"websocket/constants"
	"websocket/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ImageMetadata struct {
	ImageId        string      `json:"image_id" dynamodbav:"image_id"`
	ImageURL       string      `json:"image_URL" dynamodbav:"image_URL"`
	CategoryTagIds map[int]int `json:"category_tag_ids" dynamodbav:"category_tag_ids"`
}

type CategoryAttribute struct {
	CategoryName   string `json:"category_name" dynamodbav:"category_name"`
	CategoryNameEn string `json:"category_name_en" dynamodbav:"category_name_en"`
}

type TagAttribute struct {
	TagName       string   `json:"tag_name" dynamodbav:"tag_name"`
	TagNameEn     string   `json:"tag_name_en" dynamodbav:"tag_name_en"`
	TagNameOthers []string `json:"tag_name_others" dynamodbav:"tag_name_others"`
}

var categoryNames map[int]CategoryAttribute
var categoryTags map[int]map[int]TagAttribute

func InitQuizImageData() {
	categoryNames = make(map[int]CategoryAttribute)
	categoryTags = make(map[int]map[int]TagAttribute)

	LoadCategoryNames()
	LoadCategoryTags()
}

func LoadCategoryNames() {
	input := &dynamodb.ScanInput{
		TableName: aws.String(constants.DYNAMO_IMAGE_CATEGORY_NAMES_TABLE_NAME),
	}

	result, err := dynamoDBClient.Scan(context.TODO(), input)
	if err != nil {
		fmt.Printf("Failed to scan ImageCategoryNames: %v\n", err)
		return
	}

	for _, item := range result.Items {
		var categoryName CategoryAttribute
		categoryID, _ := strconv.Atoi(item["category_id"].(*types.AttributeValueMemberN).Value)

		err = attributevalue.UnmarshalMap(item, &categoryName)
		if err != nil {
			fmt.Printf("Failed to unmarshal CategoryAttribute: %v\n", err)
			continue
		}

		categoryNames[categoryID] = categoryName
	}
}

func LoadCategoryTags() {
	for categoryID := range categoryNames {
		categoryTags[categoryID] = make(map[int]TagAttribute)

		input := &dynamodb.QueryInput{
			TableName:              aws.String(constants.DYNAMO_IMAGE_CATEGORY_TAGS_TABLE_NAME),
			KeyConditionExpression: aws.String("category_id = :cid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":cid": &types.AttributeValueMemberN{Value: strconv.Itoa(categoryID)},
			},
		}

		result, err := dynamoDBClient.Query(context.TODO(), input)
		if err != nil {
			fmt.Printf("Failed to query ImageCategoryTags for category %d: %v\n", categoryID, err)
			continue
		}

		for _, item := range result.Items {
			var tagAttribute TagAttribute
			tagID, _ := strconv.Atoi(item["tag_id"].(*types.AttributeValueMemberN).Value)

			err = attributevalue.UnmarshalMap(item, &tagAttribute)
			if err != nil {
				fmt.Printf("Failed to unmarshal TagAttribute: %v\n", err)
				continue
			}

			categoryTags[categoryID][tagID] = tagAttribute
		}
	}
}

func GetRandomImages() ([constants.IMAGE_COUNT]string, [constants.IMAGE_COUNT][constants.CATEGORY_COUNT]models.Keyword, error) {
	var imageURLs [constants.IMAGE_COUNT]string
	var imageKeywords [constants.IMAGE_COUNT][constants.CATEGORY_COUNT]models.Keyword

	for i := 0; i < constants.IMAGE_COUNT; i++ {
		categoryIds := GetRandomCategoryIds(constants.CATEGORY_COUNT)
		var selectedTagIds []int

		for j, categoryId := range categoryIds {
			tagId := GetRandomTagId(categoryId)
			selectedTagIds = append(selectedTagIds, tagId)

			imageKeywords[i][j] = models.Keyword{
				CategoryId:   categoryId,
				CategoryName: categoryNames[categoryId].CategoryName,
				TagId:        tagId,
				TagName:      categoryTags[categoryId][tagId].TagName,
				Answers:      append([]string{categoryTags[categoryId][tagId].TagName, categoryTags[categoryId][tagId].TagNameEn}, categoryTags[categoryId][tagId].TagNameOthers...),
			}
		}

		imageURL, err := GetImageURL(categoryIds, selectedTagIds)
		if err != nil {
			return [constants.IMAGE_COUNT]string{}, [constants.IMAGE_COUNT][constants.CATEGORY_COUNT]models.Keyword{}, err
		}

		presignedURL, err := GetPresignedURLForImage(constants.S3_IMAGE_PREFIX_DIR + path.Base(imageURL))
		if err != nil {
			return [constants.IMAGE_COUNT]string{}, [constants.IMAGE_COUNT][constants.CATEGORY_COUNT]models.Keyword{}, err
		}

		imageURLs[i] = presignedURL
	}

	return imageURLs, imageKeywords, nil
}

func GetRandomCategoryIds(count int) []int {
	var categoryIds []int
	for id := range categoryNames {
		categoryIds = append(categoryIds, id)
	}

	rand.Shuffle(len(categoryIds), func(i, j int) {
		categoryIds[i], categoryIds[j] = categoryIds[j], categoryIds[i]
	})

	return categoryIds[:count]
}

func GetRandomTagId(categoryId int) int {
	var tagIds []int
	for id := range categoryTags[categoryId] {
		tagIds = append(tagIds, id)
	}

	return tagIds[rand.Intn(len(tagIds))]
}

func GetImageURL(categoryIds []int, tagIds []int) (string, error) {
	if len(categoryIds) != constants.CATEGORY_COUNT || len(tagIds) != constants.CATEGORY_COUNT {
		return "", fmt.Errorf("exactly 3 category IDs and tag IDs are required")
	}

	// Pair categoryIds and tagIds
	pairs := make([]models.CategoryTagPair, constants.CATEGORY_COUNT)
	for i := 0; i < constants.CATEGORY_COUNT; i++ {
		pairs[i] = models.CategoryTagPair{CategoryId: categoryIds[i], TagId: tagIds[i]}
	}

	// Sort pairs by CategoryId in ascending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].CategoryId < pairs[j].CategoryId
	})

	// Create image_id
	var imageIdParts []string
	for _, pair := range pairs {
		idPart := strconv.Itoa(pair.CategoryId*100 + pair.TagId)
		imageIdParts = append(imageIdParts, idPart)
	}
	imageId := imageIdParts[0] + imageIdParts[1] + imageIdParts[2]

	// Query DynamoDB using the image_id as partition key
	input := &dynamodb.GetItemInput{
		TableName: aws.String(constants.DYNAMO_IMAGE_METADATA_ASSOCIATION_TABLE_NAME),
		Key: map[string]types.AttributeValue{
			"image_id": &types.AttributeValueMemberN{Value: imageId},
		},
	}

	result, err := dynamoDBClient.GetItem(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to get item from ImageMetadataAssociation: %v", err)
	}

	if result.Item == nil {
		return "", fmt.Errorf("no image found for image_id: %s", imageId)
	}

	var imageMetadata ImageMetadata
	err = attributevalue.UnmarshalMap(result.Item, &imageMetadata)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal image metadata: %v", err)
	}

	return imageMetadata.ImageURL, nil
}
