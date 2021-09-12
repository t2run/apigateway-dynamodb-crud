package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func main() {
	lambda.Start(handler)
}

//Table Structure
type UserInfo struct {

	// User Id  user
	UserId string `json:"userId,omitempty"`

	// First name of the logged user
	FirstName string `json:"firstName,omitempty"`

	// Last name of the logged user
	LastName string `json:"lastName,omitempty"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	switch request.HTTPMethod {
	case "GET":
		userId := request.PathParameters["userId"]
		userDetails, err := GetUser(userId)
		if err != nil {
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}
		body := toString(userDetails)
		return generateAPIResponse(http.StatusOK, body), nil

	case "POST":
		var user UserInfo
		errUnMarshal := json.Unmarshal([]byte(request.Body), &user)
		if errUnMarshal != nil {
			fmt.Println("[Error] ", errUnMarshal.Error())
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}
		_, errCreateUser := CreateNewUser(user)
		if errCreateUser != nil {
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}

		return generateAPIResponse(http.StatusCreated, ""), nil

	case "PUT":
		var user UserInfo
		errUnMarshal := json.Unmarshal([]byte(request.Body), &user)
		if errUnMarshal != nil {
			fmt.Println("[Error] ", errUnMarshal.Error())
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}
		_, errUpdateUser := UpdateUserInfo(user)
		if errUpdateUser != nil {
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}
		return generateAPIResponse(http.StatusCreated, ""), nil

	case "DELETE":

		userId := request.PathParameters["userId"]
		err := DeleteUser(userId)
		if err != nil {
			return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
		}
		return generateAPIResponse(http.StatusOK, ""), nil

	default:
		return generateAPIResponse(http.StatusInternalServerError, "Internal Error"), errors.New("Internal Error")
	}
}

func toString(user UserInfo) (userString string) {
	out, _ := json.Marshal(user)
	return string(out)
}

func generateAPIResponse(code int, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: code,
	}
}

//GetUsersTemplate details
func GetUser(userID string) (UserInfo, error) {

	var userInfo UserInfo
	region := "ap-south-1"
	tableName := "UserInfo"
	awsSession, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	dynaClient := dynamodb.New(awsSession)

	keys := make(map[string]*dynamodb.AttributeValue)
	itemKeyValue := dynamodb.AttributeValue{S: aws.String(userID)}
	//Primary key
	keys["userId"] = &itemKeyValue
	getItemInput := dynamodb.GetItemInput{TableName: aws.String(tableName), Key: keys}
	response, errFromLookup := dynaClient.GetItem(&getItemInput)
	if errFromLookup != nil {
		errorString := "FailedTableLookupError" + "[" + errFromLookup.Error() + "]"
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}
	if response.Item == nil {
		errorString := "UserNotFound" + ": " + userID
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}
	errFromItemUnmarshal := dynamodbattribute.UnmarshalMap(response.Item, &userInfo)
	if errFromItemUnmarshal != nil {
		errorString := "ItemUnMarshalError" + ": " + userID
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}

	fmt.Println(" User details of userID : " + userID + " Fetched Successfully")
	return userInfo, nil
}

//CreateNewUser
func CreateNewUser(userInfo UserInfo) (UserInfo, error) {
	var user UserInfo
	userTableName := "UserInfo"
	region := "ap-south-1"
	awsSession, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	dynaClient := dynamodb.New(awsSession)

	// Save Users
	inputItemValue, errMarshalMap := dynamodbattribute.MarshalMap(userInfo)
	if errMarshalMap != nil {
		errorString := "Marshal Map Error" + "[" + errMarshalMap.Error() + "]"
		fmt.Println(errorString)
		return user, errors.New(errorString)
	}

	input := &dynamodb.PutItemInput{
		Item:      inputItemValue,
		TableName: aws.String(userTableName),
	}

	_, errPutItem := dynaClient.PutItem(input)
	if errPutItem != nil {
		errorString := "Put Item Error" + "[" + errPutItem.Error() + "]"
		fmt.Println(errorString)
		return user, errors.New(errorString)
	}
	fmt.Println("User : " + user.UserId + " Created Successfully")
	return user, nil
}

//UpdateUserInfo in DynamoDB Users Details
func UpdateUserInfo(userInfo UserInfo) (UserInfo, error) {

	region := "ap-south-1"
	userTableName := "UserInfo"
	awsSession, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	dynaClient := dynamodb.New(awsSession)

	//UserInfoUpdate model
	type UserInfoUpdate struct {
		FirstName string `json:":firstName,omitempty"`
		LastName  string `json:":lastName,omitempty"`
	}

	//RoleInfoKey model
	type RoleInfoKey struct {
		RoleName string `json:"roleName"`
	}

	//UserInfoItemKey model
	type UserInfoKey struct {
		UserID string `json:"userId"`
	}

	updateUserInfo := UserInfoUpdate{
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
	}

	av, KeyErr := dynamodbattribute.MarshalMap(UserInfoKey{UserID: userInfo.UserId})
	if KeyErr != nil {
		errorString := "FailedTableLookupError" + "[" + KeyErr.Error() + "]"
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}

	updateDetails, errUpdateDetails := dynamodbattribute.MarshalMap(updateUserInfo)
	if errUpdateDetails != nil {
		errorString := "FailedToCreateUpdateDetails" + "[" + errUpdateDetails.Error() + "]"
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}

	input := &dynamodb.UpdateItemInput{
		Key:       av,
		TableName: aws.String(userTableName),
		// ExpressionAttributeNames:  map[string]*string{"#role": aws.String("role")},
		UpdateExpression:          aws.String("set firstName = :firstName, lastName = :lastName"),
		ExpressionAttributeValues: updateDetails,
	}

	_, errUpdateItem := dynaClient.UpdateItem(input)
	if errUpdateItem != nil {
		errorString := "UpdateItemError" + "[" + errUpdateItem.Error() + "]"
		fmt.Println(errorString)
		return userInfo, errors.New(errorString)
	}

	fmt.Println("User : " + userInfo.UserId + " Details Updated Successfully")
	return userInfo, nil
}

//DeleteUser from the table
func DeleteUser(userID string) error {
	region := "ap-south-1"
	userTableName := "UserInfo"
	awsSession, _ := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	dynaClient := dynamodb.New(awsSession)

	keys := make(map[string]*dynamodb.AttributeValue)
	itemKeyValue := dynamodb.AttributeValue{S: aws.String(userID)}
	keys["userId"] = &itemKeyValue

	deleteItemInput := dynamodb.DeleteItemInput{TableName: aws.String(userTableName), Key: keys}
	_, errFromDelete := dynaClient.DeleteItem(&deleteItemInput)
	if errFromDelete != nil {
		errorString := "Failed to Delete" + "[" + errFromDelete.Error() + "]"
		fmt.Println(errorString)
		return errors.New(errorString)
	}
	fmt.Println("User : " + userID + " Deleted Successfully")
	return nil
}
