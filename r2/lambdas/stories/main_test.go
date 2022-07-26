package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/sowens-csd/folktells-server/ftdb"
	"github.com/sowens-csd/folktells-server/sharing"
)

func TestSingleStoryReturnedAsJSON(t *testing.T) {
	stories := handleSuccessfulCall("GroupId1", t)
	story := stories.Stories[0]
	if story.StoryID != "story1" {
		t.Errorf("Expected content not found")
	}
}

func TestMultipleStoriesReturnedAsJSON(t *testing.T) {
	stories := handleSuccessfulCall("GroupId2", t)
	if 2 != len(stories.Stories) {
		t.Errorf("Expected 2, was %d", len(stories.Stories))
	}
	story1 := stories.Stories[0]
	if story1.StoryID != "story1" {
		t.Errorf("Expected content not found")
	}
	story2 := stories.Stories[1]
	if story2.StoryID != "story2" {
		t.Errorf("Second story ID was %s", story2.StoryID)
	}
}

func TestNoStoriesReturnedAsEmptyJSON(t *testing.T) {
	stories := handleSuccessfulCall("groupIdNotFound", t)
	if len(stories.Stories) != 0 {
		t.Errorf("Stories not empty")
	}
}

func handleSuccessfulCall(groupID string, t *testing.T) sharing.SharedStories {
	svc := &stubDynamoDB{}
	requestLogger := log.WithFields(log.Fields{"request_id": "Test", "group_id": groupID})
	resp := storiesForGroup(groupID, svc, requestLogger)
	if resp.StatusCode != 200 {
		t.Errorf("Got a non nil error.")
	}
	stories := parseStoriesJSON(resp.Body, t)
	return stories
}

func parseStoriesJSON(body string, t *testing.T) sharing.SharedStories {
	var stories sharing.SharedStories
	storiesJSON := []byte(body)
	err := json.Unmarshal(storiesJSON, &stories)
	if err != nil {
		t.Errorf(err.Error())
	}
	return stories
}

type stubDynamoDB struct {
	dynamodbiface.DynamoDBAPI
}

func (m *stubDynamoDB) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	// Returned canned response
	var count int64
	if *input.ExpressionAttributeValues[":resId"].S == "G#GroupId1" {
		count = 1
	} else if *input.ExpressionAttributeValues[":resId"].S == "G#GroupId2" {
		count = 2
	}
	var items []map[string]*dynamodb.AttributeValue
	if count > 0 {
		for result := 0; int64(result) < count; result++ {
			var item map[string]*dynamodb.AttributeValue = make(map[string]*dynamodb.AttributeValue)
			storyID := fmt.Sprintf("story%d", result+1)
			item[ftdb.ResourceIDField] = &dynamodb.AttributeValue{
				S: aws.String("G#GroupId1"),
			}
			item[ftdb.ReferenceIDField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("S#story%d", result+1)),
			}
			item[ftdb.IDField] = &dynamodb.AttributeValue{
				S: aws.String(storyID),
			}
			item[ftdb.ContentField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("# Story %d\nSome sample story content", result+1)),
			}
			item[ftdb.VersionField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("version%d", result+1)),
			}
			item[ftdb.BaseVersionField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("baseVersion%d", result+1)),
			}
			item[ftdb.LastUpdatedField] = &dynamodb.AttributeValue{
				N: aws.String("1583440478299"),
			}
			item[ftdb.LastUpdatedByField] = &dynamodb.AttributeValue{
				S: aws.String(fmt.Sprintf("user%d", result+1)),
			}
			item[ftdb.StorySourceField] = &dynamodb.AttributeValue{
				S: aws.String("googlePhotos"),
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
