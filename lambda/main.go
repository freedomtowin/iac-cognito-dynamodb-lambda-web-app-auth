package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	math_rand "math/rand"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Request struct {
	Operation string      `json:"operation"`
	Payload   interface{} `json:"payload"`
}

type User struct {
	UserID       string  `json:"user_id"`
	Email        string  `json:"email"`
	WalletAmount float64 `json:"wallet_amount"`
}

type ApiKey struct {
	UserID string `json:"user_id"`
	ApiKey string `json:"api_key"`
}

type Transaction struct {
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}

var svc *dynamodb.DynamoDB

func init() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Fatalf("unable to create AWS session, %v", err)
	}
	svc = dynamodb.New(sess)
}

func handler(ctx context.Context, request Request) (string, error) {
	switch request.Operation {
	case "createUser":
		user := request.Payload.(map[string]interface{})
		return createUser(user)
	case "getUser":
		userID := request.Payload.(string)
		return getUser(userID)
	//Admin Only
	case "updateWallet":
		payload := request.Payload.(map[string]interface{})
		return updateWallet(payload["user_id"].(string), payload["amount"].(float64))
	case "addWallet":
		payload := request.Payload.(map[string]interface{})
		return addWallet(payload["user_id"].(string), payload["amount"].(float64))
	// Used by the front end application to display api keys
	case "getApiKeyFromUser":
		userID := request.Payload.(string)
		return getApiKeyFromUser(userID)
	// Use by the service that the API key is used for
	case "getUserFromApiKey":
		apiKey := request.Payload.(string)
		return getUserFromApiKey(apiKey)
	case "generateApiKey":
		userID := request.Payload.(string)
		return generateApiKey(userID)
	case "logTransaction":
		transaction := request.Payload.(map[string]interface{})
		return logTransaction(transaction)
	case "getTransactionHistory":
		userID := request.Payload.(string)
		return getTransactionHistory(userID)
	case "callAPI":
		apiKey := request.Payload.(string)
		return callAPI(apiKey)
	default:
		return "", fmt.Errorf("invalid operation")
	}
}

func createUser(user map[string]interface{}) (string, error) {
	userItem, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user, %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("users"),
		Item:      userItem,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to create user, %v", err)
	}

	return "User created successfully", nil
}

func getUser(userID string) (string, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to get user, %v", err)
	}

	if result.Item == nil {
		return "", fmt.Errorf("user not found")
	}

	user := User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal user, %v", err)
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user JSON, %v", err)
	}

	return string(userJson), nil
}

func getApiKeyFromUser(userID string) (string, error) {

	indexName := "user_id-index" // Make sure this is the correct index name

	input := &dynamodb.QueryInput{
		TableName: aws.String("api_keys"),
		IndexName: aws.String(indexName),
		KeyConditions: map[string]*dynamodb.Condition{
			"user_id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userID),
					},
				},
			},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return "", fmt.Errorf("failed to get API key, %v", err)
	}

	apiKeys := []ApiKey{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &apiKeys)
	if err != nil {
		return "", fmt.Errorf("API key not found")
	}

	apiKeyJson, err := json.Marshal(apiKeys[0])
	if err != nil {
		return "", fmt.Errorf("failed to marshal API key JSON, %v", err)
	}

	return string(apiKeyJson), nil
}

func getUserFromApiKey(apiKey string) (string, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String("api_keys"),
		Key: map[string]*dynamodb.AttributeValue{
			"api_key": {
				S: aws.String(apiKey),
			},
		},
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to get API key, %v", err)
	}

	if result.Item == nil {
		return "", fmt.Errorf("API key not found")
	}

	apiKeyData := ApiKey{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &apiKeyData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal API key, %v", err)
	}

	return apiKeyData.UserID, nil
}

func addWallet(userID string, amount float64) (string, error) {

	if amount <= 0 {
		return "", fmt.Errorf("invalid wallet amount")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
		UpdateExpression: aws.String("SET wallet_amount = wallet_amount + :amount"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":amount": {
				N: aws.String(fmt.Sprintf("%f", amount)),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to update wallet amount, %v", err)
	}

	return "Wallet amount updated successfully", nil
}

func updateWallet(userID string, amount float64) (string, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String("users"),
		Key: map[string]*dynamodb.AttributeValue{
			"user_id": {
				S: aws.String(userID),
			},
		},
		UpdateExpression: aws.String("SET wallet_amount = wallet_amount + :amount"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":amount": {
				N: aws.String(fmt.Sprintf("%f", amount)),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to update wallet amount, %v", err)
	}

	return "Wallet amount updated successfully", nil
}

func generateApiKey(userID string) (string, error) {

	keyHolder := make([]byte, 32) // 32 bytes will be 256 bits

	// Read random bytes into the byte slice
	_, err := rand.Read(keyHolder)
	if err != nil {
		return "", err
	}

	apiKey := base64.StdEncoding.EncodeToString(keyHolder)
	keyItem := ApiKey{
		UserID: userID,
		ApiKey: apiKey,
	}

	keyItemMap, err := dynamodbattribute.MarshalMap(keyItem)
	if err != nil {
		return "", fmt.Errorf("failed to marshal API key, %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("api_keys"),
		Item:      keyItemMap,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to store API key, %v", err)
	}

	return apiKey, nil
}

func logTransaction(transaction map[string]interface{}) (string, error) {

	transactionID := time.Now().UnixNano() / int64(time.Millisecond) // Milliseconds since epoch

	// Add transaction ID to the transaction map
	transaction["transaction_id"] = transactionID

	transactionItem, err := dynamodbattribute.MarshalMap(transaction)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction, %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("transactions"),
		Item:      transactionItem,
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return "", fmt.Errorf("failed to log transaction, %v", err)
	}

	return "Transaction logged successfully", nil
}

func getTransactionHistory(userID string) (string, error) {

	indexName := "user_id-index" // Make sure this is the correct index name

	input := &dynamodb.QueryInput{
		TableName: aws.String("transactions"),
		IndexName: aws.String(indexName),
		KeyConditions: map[string]*dynamodb.Condition{
			"user_id": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(userID),
					},
				},
			},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return "", fmt.Errorf("failed to query transaction history, %v", err)
	}

	transactions := []Transaction{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &transactions)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal transactions, %v", err)
	}

	transactionsJson, err := json.Marshal(transactions)
	if err != nil {
		return "", fmt.Errorf("failed to marshal transactions JSON, %v", err)
	}

	return string(transactionsJson), nil
}

func callAPI(apiKey string) (string, error) {

	userID, err := getUserFromApiKey(apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to get user, %v", err)
	}

	// Generate a random number between 0 and 1
	randomDuration := float64(math_rand.Intn(1000)) / 1000.0

	// Convert the random duration to milliseconds and sleep for that duration
	sleepDuration := time.Duration(randomDuration * float64(time.Second))
	time.Sleep(sleepDuration)

	_, err_update := updateWallet(userID, -1*randomDuration)

	if err_update != nil {
		return "", fmt.Errorf("failed to updated wallet, %v", err_update)
	}

	data := make(map[string]interface{})
	description := "api call cost"
	data["user_id"] = userID
	data["amount"] = -1 * randomDuration
	data["description"] = description

	_, err_transaction := logTransaction(data)

	if err_transaction != nil {
		return "", fmt.Errorf("failed to log transaction, %v", err_transaction)
	}
	// Return the sleep duration in milliseconds as a string

	return strconv.FormatFloat(randomDuration*1000, 'f', -1, 64) + " ms", nil
}

func main() {
	lambda.Start(handler)
}
