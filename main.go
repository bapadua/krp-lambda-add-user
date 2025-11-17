package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type Request struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email,omitempty"`
	Age         int    `json:"age,omitempty"`
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email,omitempty"`
	Age         int    `json:"age,omitempty"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

var dynamoClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func HandleRequest(ctx context.Context, req Request) (Response, error) {
	// Validar entrada
	if req.Name == "" {
		return errorResponse(400, "name is required")
	}
	if req.PhoneNumber == "" {
		return errorResponse(400, "phone_number is required")
	}

	// Criar usuário
	user := User{
		ID:          uuid.New().String(),
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		Age:         req.Age,
		Status:      "active",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	// Preparar item para DynamoDB
	item := map[string]types.AttributeValue{
		"id":           &types.AttributeValueMemberS{Value: user.ID},
		"phone_number": &types.AttributeValueMemberS{Value: user.PhoneNumber},
		"name":         &types.AttributeValueMemberS{Value: user.Name},
		"status":       &types.AttributeValueMemberS{Value: user.Status},
		"created_at":   &types.AttributeValueMemberS{Value: user.CreatedAt},
	}

	if user.Email != "" {
		item["email"] = &types.AttributeValueMemberS{Value: user.Email}
	}

	if user.Age > 0 {
		item["age"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", user.Age)}
	}

	// Salvar no DynamoDB
	tableName := os.Getenv("USERS_TABLE_NAME")
	if tableName == "" {
		return errorResponse(500, "USERS_TABLE_NAME environment variable not set")
	}

	_, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return errorResponse(500, fmt.Sprintf("failed to save user: %v", err))
	}

	// Resposta de sucesso
	responseBody, _ := json.Marshal(map[string]interface{}{
		"message": "User created successfully",
		"user":    user,
	})

	return Response{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func errorResponse(code int, msg string) (Response, error) {
	body, _ := json.Marshal(map[string]string{"error": msg})
	return Response{
		StatusCode: code,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

