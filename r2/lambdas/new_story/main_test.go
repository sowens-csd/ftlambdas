package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

func TestSingleStoryReturnedAsJSON(t *testing.T) {
	resp := handleSuccessfulCall("GroupId1", t)
	if resp.StatusCode != 200 {
		t.Errorf("Expected response not found %s", resp.Body)
	}
}

func handleSuccessfulCall(groupID string, t *testing.T) awsproxy.Response {
	svc := &stubDynamoDB{}
	resp := shareStoryToGroup("{\"groupId\":\"group1\",\"sharedStory\":{\"id\":\"story1\",\"content\":\"content1\",\"version\":\"v1\",\"baseVersion\":\"v2\",\"lastUpdated\":84893209,\"lastUpdatedBy\":\"author1\",\"storySource\":\"googlePhotos\"}}", "requestID1", svc)
	return resp
}

type stubDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}

func (m *stubDynamoDB) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	if input.Key[ftdb.ResourceIDField] == nil || input.Key[ftdb.ReferenceIDField] == nil {
		return nil, errors.New("Nil key component")
	}
	resID := *input.Key[ftdb.ResourceIDField].S
	refID := *input.Key[ftdb.ReferenceIDField].S
	if resID != "G#group1" && refID != "S#story1" {
		return nil, fmt.Errorf("Resource or referenceId didn't match, %s, %s", resID, refID)
	}
	return &dynamodb.UpdateItemOutput{}, nil
}
