package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sowens-csd/folktells-server/ftdb"
)

const request1 = "request1"
const email1 = "user1@example.com"
const email2 = "user2@example.com"
const user1 = "user1"
const userResourceID1 = "U#" + user1
const user2 = "user2"
const invitedOn1 = 29304802

func TestFindsSingleUser(t *testing.T) {
	svc := &stubDynamoDB{}
	resp := findUser(email1, "request1", svc)
	if resp.StatusCode != 200 {
		t.Errorf("Bad response code")
	}
}

func TestReturns404OnFailToFindsUser(t *testing.T) {
	svc := &stubDynamoDB{}
	resp := findUser(email2, "request1", svc)
	if resp.StatusCode != 404 {
		t.Errorf("Expected a 404")
	}
}

type stubDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}

func (m *stubDynamoDB) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	// Returned canned response
	var count int64
	if *input.ExpressionAttributeValues[":email"].S == email1 {
		count = 1
	} else if *input.ExpressionAttributeValues[":email"].S == "multipleEmail" {
		count = 2
	} else {
		count = 0
	}
	var items []map[string]*dynamodb.AttributeValue
	if count > 0 {
		for result := 0; int64(result) < count; result++ {
			var item map[string]*dynamodb.AttributeValue = make(map[string]*dynamodb.AttributeValue)
			userResourceID := fmt.Sprintf("U#user%d", result+1)
			item[ftdb.ResourceIDField] = &dynamodb.AttributeValue{
				S: aws.String(userResourceID),
			}
			item[ftdb.IDField] = &dynamodb.AttributeValue{
				S: aws.String(user1),
			}
			item[ftdb.EmailField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("email%d", result+1)),
			}
			item[ftdb.CreatedAtField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("4329447")),
			}
			items = append(items, item)
		}
	}
	output := dynamodb.QueryOutput{
		Count: &count,
		Items: items,
	}
	// output.Count = &count
	return &output, nil
}
