package sharing

import (
	"encoding/json"
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
)

var storyTestDBData = awsproxy.TestDBData{
	testUser1Group1MembershipRecord(),
	testUser1Record(),
	testStory1Record(),
	testStory1Group1Record(),
}

func TestStoryAndGroupRecordsUpdated(t *testing.T) {
	testDB := awsproxy.NewTestDBSvc()
	ftCtx := awsproxy.NewTestContext(userID2, testDB)
	inputJSON := []byte(shareStoryJSON1)
	var sharedStory SharedStory
	err := json.Unmarshal(inputJSON, &sharedStory)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	}
	_, err = UpdateSharedStory(ftCtx, sharedStory)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	}
}

func TestDuplicateStoryBlockedOnUpdate(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(storyTestDBData)
	ftCtx := awsproxy.NewTestContext(userID2, testDB)
	inputJSON := []byte(duplicateShareStoryJSON1)
	var sharedStory SharedStory
	err := json.Unmarshal(inputJSON, &sharedStory)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	}
	_, err = UpdateSharedStory(ftCtx, sharedStory)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	}
	testDB.ExpectPutCount(0, t)
}

func TestLoadPopulatesStory(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(storyTestDBData)
	ftCtx := awsproxy.NewTestContext(userID2, testDB)
	story, err := LoadSharedStory(ftCtx, storyID1)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	} else {
		if story.StoryID != storyID1 {
			t.Errorf("Expected %s, got %s", storyID1, story.StoryID)
		}
		if story.Content != content1 {
			t.Errorf("Expected %s, got %s", content1, story.Content)
		}
	}
}

func TestFindsSingleSharedStoryForUser(t *testing.T) {
	testDB := awsproxy.NewTestDBSvcWithData(storyTestDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testDB)
	sharedStories, err := FindSharedStoriesForUser(ftCtx)
	if nil != err {
		t.Errorf("Failed %s", err.Error())
	} else {
		if len(sharedStories.Stories) != 1 {
			t.Errorf("Should have had 1 story was %d", len(sharedStories.Stories))
		} else {
			story := sharedStories.Stories[0]
			if story.Content != content1 {
				t.Errorf("Expected %s, got %s", content1, story.Content)
			}
		}
	}
}
