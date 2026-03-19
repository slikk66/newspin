package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Client struct {
	dynamo         *dynamodb.Client
	usersTableName string
	pinsTableName  string
}

type User struct {
	Username     string `json:"username" dynamodbav:"username"`
	PasswordHash string `json:"-" dynamodbav:"passwordHash"`
	CreatedAt    string `json:"createdAt" dynamodbav:"createdAt"`
}

type Pin struct {
	UserId    string `json:"userId" dynamodbav:"userId"`
	ArticleId string `json:"articleId" dynamodbav:"articleId"`
	Title     string `json:"title" dynamodbav:"title"`
	Url       string `json:"url" dynamodbav:"url"`
	PinnedAt  string `json:"pinnedAt" dynamodbav:"pinnedAt"`
}

func NewClient(usersTable, pinsTable string) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	return &Client{
		dynamo:         dynamodb.NewFromConfig(cfg),
		usersTableName: usersTable,
		pinsTableName:  pinsTable,
	}, nil
}

func (c *Client) CreateUser(username, passwordHash string) error {
	item, err := attributevalue.MarshalMap(User{
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	_, err = c.dynamo.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName:           &c.usersTableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(username)"),
	})
	return err
}

func (c *Client) GetUser(username string) (*User, error) {

	result, err := c.dynamo.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: &c.usersTableName,
		Key: map[string]types.AttributeValue{
			"username": &types.AttributeValueMemberS{Value: username},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("no user found!")
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) CreatePin(pin *Pin) error {
	pin.PinnedAt = time.Now().UTC().Format(time.RFC3339)
	item, err := attributevalue.MarshalMap(pin)

	if err != nil {
		return err
	}
	_, err = c.dynamo.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: &c.pinsTableName,
		Item:      item,
	})
	return err
}

func (c *Client) GetPins(userId string) ([]Pin, error) {

	result, err := c.dynamo.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              &c.pinsTableName,
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userId},
		},
	})
	if err != nil {
		return nil, err
	}
	var pins []Pin
	err = attributevalue.UnmarshalListOfMaps(result.Items, &pins)
	if err != nil {
		return nil, err
	}
	return pins, nil
}

func (c *Client) DeletePin(userId, articleId string) error {

	_, err := c.dynamo.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: &c.pinsTableName,
		Key: map[string]types.AttributeValue{
			"userId":    &types.AttributeValueMemberS{Value: userId},
			"articleId": &types.AttributeValueMemberS{Value: articleId},
		},
	})
	return err

}
