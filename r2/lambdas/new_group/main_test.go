package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/ftdb"
)

func TestGroupAdded(t *testing.T) {
	resp := handleSuccessfulCall("GroupId1", "user1", t)
	if resp.StatusCode != 200 {
		t.Errorf("Expected response not found %s", resp.Body)
	}
}

func handleSuccessfulCall(groupID string, userID string, t *testing.T) awsproxy.Response {
	svc := &stubDynamoDB{}
	resp := createGroup("{\"id\":\"group1\",\"name\":\"name1\"}", "requestID1", userID, svc)
	return resp
}

type stubDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}

func (m *stubDynamoDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	// Make response
	resourceID := dynamodb.AttributeValue{}
	resourceID.SetS("U#user1")
	email := dynamodb.AttributeValue{}
	email.SetS("user@example.com")
	resp := make(map[string]*dynamodb.AttributeValue)
	resp[ftdb.ResourceIDField] = &resourceID
	resp[ftdb.ReferenceIDField] = &resourceID
	resp[ftdb.EmailField] = &email

	// Returned canned response
	output := &dynamodb.GetItemOutput{
		Item: resp,
	}
	return output, nil
}

func (m *stubDynamoDB) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if input.Item[ftdb.ResourceIDField] == nil || input.Item[ftdb.ReferenceIDField] == nil {
		return nil, errors.New("Nil key component")
	}
	resID := *input.Item[ftdb.ResourceIDField].S
	refID := *input.Item[ftdb.ReferenceIDField].S
	if resID != "G#group1" && refID != "G#group1" {
		return nil, fmt.Errorf("Resource or referenceId didn't match, %s, %s", resID, refID)
	}
	return &dynamodb.PutItemOutput{}, nil
}
