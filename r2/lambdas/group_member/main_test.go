package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sowens-csd/folktells-server/ftdb"
	"github.com/sowens-csd/folktells-server/sharing"
)

const request1 = "request1"
const group1 = "group1"
const group2 = "group2"
const inviteID1 = "invite1"
const inviteID2 = "invite2"
const email1 = "user1@example.com"
const email2 = "user2@example.com"
const newUserEmail = "new.user@example.com"
const user1 = "user1"
const userResourceID1 = "U#" + user1
const user2 = "user2"
const invitedOn1 = 29304802

func TestInviteFromJson(t *testing.T) {
	inviteJSON := `
		{
			"invitationId":"787b5910-f7f8-5614-b973-339fbd99e141",
			"groupId":"group1",
			"memberEmail":"user1@example.com",
			"invitedById":"787b5910-f7f8-5614-b973-339fbd99e143",
			"invitedOn": 29304802
		}
	`
	svc := &stubDynamoDB{}
	resp := updateGroupMember(inviteJSON, user1, request1, svc)
	if resp.StatusCode != 200 {
		t.Errorf("Failed response %s", resp.Body)
	}
}

func TestErrorIfMultipleUsersFound(t *testing.T) {
	svc := &stubDynamoDB{}
	requestLogger := log.WithFields(log.Fields{})
	_, err := findInvitedUser("multipleEmail", svc, requestLogger)
	if nil == err {
		t.Errorf("No error on multiple users")
	}
}

func TestInviteInsertedWhenNone(t *testing.T) {
	svc := &stubDynamoDB{}
	requestLogger := log.WithFields(log.Fields{})
	groupInvite := sharing.GroupMember{
		GroupID:        group1,
		MemberEmail:    email1,
		InvitedByID:    user2,
		InvitedByEmail: user2,
	}
	_, err := insertInviteIntoDb(userResourceID1, email1, groupInvite, svc, requestLogger)
	if nil != err {
		t.Errorf("Failed to insert")
	}
}

func TestInviteBlockedWhenPresent(t *testing.T) {
	svc := &stubDynamoDB{}
	requestLogger := log.WithFields(log.Fields{})
	groupInvite := sharing.GroupMember{
		GroupID:        group2,
		MemberEmail:    email1,
		InvitedByID:    user2,
		InvitedByEmail: user2,
	}
	_, err := insertInviteIntoDb(user1, email1, groupInvite, svc, requestLogger)
	if nil == err {
		t.Errorf("Should have failed to insert due to conditional")
	}
}

type stubDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}

func (m *stubDynamoDB) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	if input.Item[ftdb.ResourceIDField] == nil || input.Item[sharing.ReferenceIDField] == nil {
		return nil, errors.New("Nil key component")
	}
	resID := *input.Item[ftdb.ResourceIDField].S
	refID := *input.Item[ftdb.ReferenceIDField].S
	if refID != ftdb.ReferenceIDFromUserID(user1) {
		return nil, fmt.Errorf("Resource didn't match, %s, %s", resID, refID)
	}
	if resID == ftdb.ResourceIDFromGroupID(group1) {
		return &dynamodb.PutItemOutput{}, nil
	} else if resID == ftdb.ResourceIDFromGroupID(group2) {
		return nil, fmt.Errorf("Conditional check failed")
	}
	return nil, fmt.Errorf("Resource didn't match, %s, %s", resID, refID)
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
			item[ftdb.EmailField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("email%d", result+1)),
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

func (m *stubDynamoDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	// Make response
	resourceID := dynamodb.AttributeValue{}
	resourceID.SetS("U#user1")
	email := dynamodb.AttributeValue{}
	email.SetS(email1)
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
