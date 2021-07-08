package sharing

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// StoryGroupActive is the value for StoryGroup.Status when the story should be
// shared with that group.
const StoryGroupActive = "y"

// StoryGroupRemoved is the value for StoryGroup.Status when the story should not
// be shared with that group.
const StoryGroupRemoved = "r"

// StoryGroup holds the relationship between one story and one group.
// It defines a shares with relationship that states that story is
// shared with that group.
type StoryGroup struct {
	StoryID       string `json:"storyId" dynamodbav:"storyId"`
	GroupID       string `json:"groupId" dynamodbav:"groupId"`
	Version       string `json:"version" dynamodbav:"version"`
	BaseVersion   string `json:"baseVersion" dynamodbav:"baseVersion"`
	LastUpdated   int    `json:"lastUpdated" dynamodbav:"lastUpdated"`
	LastUpdatedBy string `json:"lastUpdatedBy" dynamodbav:"lastUpdatedBy"`
	Status        string `json:"status" dynamodbav:"status"`
}

// SharedStory is the content for a single story
type SharedStory struct {
	StoryID        string       `json:"id" dynamodbav:"id"`
	AlbumReference string       `json:"albumReference" dynamodbav:"albumReference"`
	Content        string       `json:"content,omitempty" dynamodbav:"content"`
	Version        string       `json:"version" dynamodbav:"version"`
	BaseVersion    string       `json:"baseVersion" dynamodbav:"baseVersion"`
	LastUpdated    int          `json:"lastUpdated" dynamodbav:"lastUpdated"`
	LastUpdatedBy  string       `json:"lastUpdatedBy" dynamodbav:"lastUpdatedBy"`
	StorySource    string       `json:"storySource" dynamodbav:"storySource"`
	Groups         []StoryGroup `json:"groups,omitempty" dynamodbav:"groups,omitempty"`
}

// SharedStories is the content for a set of stories
type SharedStories struct {
	PageToken string        `json:"pageToken"`
	Stories   []SharedStory `json:"stories"`
}

// StoryNotFoundError returned by LoadSharedStory when the story is not in the DB
type StoryNotFoundError struct {
	error
	StoryID string
}

// StoryUpdateResult Returned by the UpdateStory call to indicate if the story
// update was successful
type StoryUpdateResult struct {
	Success          bool   `json:"success"`
	Status           int    `json:"status"`
	DuplicateStoryID string `json:"storyId"`
}

// DeleteStoriesForUser removes all stories from groups where this user is the only member
func DeleteStoriesForUser(ctx awsproxy.FTContext) error {
	return nil
}

func findUniqueStoriesForUser(ctx awsproxy.FTContext) (map[string][]*SharedStory, error) {
	groups, err := FindGroupsForUser(ctx)
	if err != nil {
		return nil, err
	}
	storyIDs, err := FindStoriesForGroups(ctx, groups)
	if err != nil {
		return nil, err
	}
	uniqueStories := make(map[string][]*SharedStory)
	for _, storyID := range storyIDs {
		ctx.RequestLogger.Debug().Str("story_id", storyID).Msg("unique, loading story")
		story, err := LoadSharedStory(ctx, storyID)
		if nil == err {
			stories, exists := uniqueStories[story.SourceAlbumReference()]
			if !exists {
				stories = make([]*SharedStory, 0)
				uniqueStories[story.SourceAlbumReference()] = stories
			}
			ctx.RequestLogger.Debug().Str("story_id", storyID).Str("ref", story.SourceAlbumReference()).Msg("adding unique story")
			uniqueStories[story.SourceAlbumReference()] = append(stories, story)
		} else {
			ctx.RequestLogger.Info().Str("story_id", storyID).Msg("Load failed")
		}
	}
	return uniqueStories, nil
}

// FindSharedStoriesForUser returns the complete set of stories available to the
// current user based on their group memberhips and groups the story is shared with.
func FindSharedStoriesForUser(ctx awsproxy.FTContext) (*SharedStories, error) {
	uniqueStories, err := findUniqueStoriesForUser(ctx)
	if nil != err {
		return nil, err
	}
	var stories []SharedStory
	for _, storyList := range uniqueStories {
		bestStory := chooseBestStory(storyList)
		ctx.RequestLogger.Info().Str("story", bestStory.StoryID).Msg("user has story")
		stories = append(stories, bestStory)
	}
	sharedStories := SharedStories{
		PageToken: "",
		Stories:   stories,
	}
	return &sharedStories, nil
}

func chooseBestStory(stories []*SharedStory) SharedStory {
	var mostRecent int = 0
	var longestContent int = 0
	var best SharedStory
	for _, story := range stories {
		if len(story.Content) > longestContent ||
			(len(story.Content) > 0 && story.LastUpdated > mostRecent) ||
			(0 == mostRecent && 0 == longestContent) {
			best = *story
			longestContent = len(story.Content)
			mostRecent = story.LastUpdated
		}
	}
	return best
}

// FindStoriesForGroups returns an array of story IDs for the stories that are shared
// with the groups.
func FindStoriesForGroups(ctx awsproxy.FTContext, groups []string) ([]string, error) {
	var stories []string
	for _, groupID := range groups {
		result, err := ctx.DBSvc.Query(ctx.Context, &dynamodb.QueryInput{
			TableName:              aws.String(ftdb.GetTableName()),
			KeyConditionExpression: aws.String("resourceId = :groupID "),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":groupID": &types.AttributeValueMemberS{Value: ftdb.ResourceIDFromGroupID(groupID)},
			},
		})
		storyGroupResults := []StoryGroup{}
		if err != nil {
			return stories, err
		} else if result.Count == 0 {
			return stories, nil
		}
		err = attributevalue.UnmarshalListOfMaps(result.Items, &storyGroupResults)
		if nil != err {
			return nil, err
		}
		for _, storyGroup := range storyGroupResults {
			if len(storyGroup.StoryID) > 0 {
				ctx.RequestLogger.Debug().Str("group", groupID).Str("story", storyGroup.StoryID).Msg("story for group")
				stories = append(stories, storyGroup.StoryID)
			}
		}
	}
	return stories, nil
}

// LoadSharedStory returns the shared story for the given storyID
func LoadSharedStory(ctx awsproxy.FTContext, storyID string) (*SharedStory, error) {
	resourceID := ftdb.ResourceIDFromStoryID(storyID)
	referenceID := ftdb.ReferenceIDFromStoryID(storyID)
	sharedStory := SharedStory{}
	found, err := ftdb.GetItem(ctx, resourceID, referenceID, &sharedStory)
	if nil != err {
		return nil, err
	}
	if !found {
		return nil, StoryNotFoundError{StoryID: storyID}
	}
	return &sharedStory, nil
}

// UpdateSharedStory ensures that the story and story group records are properly updated given
// the JSON that is provided.
func UpdateSharedStory(ctx awsproxy.FTContext, sharedStory SharedStory) (StoryUpdateResult, error) {
	duplicateStoryID, isDuplicate := sharedStory.IsDuplicate(ctx)
	if isDuplicate {
		return StoryUpdateResult{Success: false, Status: 1, DuplicateStoryID: duplicateStoryID}, nil
	}

	if nil != sharedStory.Groups {
		for groupIndex, groupToUpdate := range sharedStory.Groups {
			err := groupToUpdate.Save(ctx)
			if nil != err {
				return StoryUpdateResult{}, err
			}
			sharedStory.Groups[groupIndex].BaseVersion = groupToUpdate.Version
		}
	}
	return sharedStory.Save(ctx)
}

// Save updates the db with the current sharedStory state
func (sharedStory *SharedStory) Save(ctx awsproxy.FTContext) (StoryUpdateResult, error) {
	duplicateStoryID, isDuplicate := sharedStory.IsDuplicate(ctx)
	if isDuplicate {
		return StoryUpdateResult{Success: false, Status: 1, DuplicateStoryID: duplicateStoryID}, nil
	}
	ctx.RequestLogger.Debug().Str("storyId", sharedStory.StoryID).Str("albumReference", sharedStory.AlbumReference).Msg("Putting shared story")
	condtionExp := " attribute_not_exists(resourceId) "
	var expressionValues map[string]types.AttributeValue = make(map[string]types.AttributeValue)
	if len(sharedStory.BaseVersion) > 0 {
		condtionExp = condtionExp + " or baseVersion=:baseVersion"
		expressionValues[":baseVersion"] = &types.AttributeValueMemberS{Value: sharedStory.BaseVersion}
	}
	sharedStory.BaseVersion = sharedStory.Version
	resourceID := ftdb.ResourceIDFromStoryID(sharedStory.StoryID)
	err := ftdb.PutItem(ctx, resourceID, resourceID, sharedStory)
	if nil != err {
		ctx.RequestLogger.Error().Msg(fmt.Sprintf("save shared story failed %s", err.Error()))
	} else {
		ctx.RequestLogger.Info().Msg("save shared story succeeded")
	}
	return StoryUpdateResult{Success: true, Status: 0}, err
}

// SourceAlbumReference returns a reference string that is unique for the same
// album reference for different sources
// This ensures that two albums titled 'Graduation' one local and one in
// Google Photos are treated as unique.
func (sharedStory *SharedStory) SourceAlbumReference() string {
	return fmt.Sprintf("%s_%s", sharedStory.StorySource, sharedStory.AlbumReference)
}

// IsDuplicate returns true if the story is a duplicate of an existing story
func (sharedStory *SharedStory) IsDuplicate(ctx awsproxy.FTContext) (string, bool) {
	isDuplicate := false
	duplicateStory := ""
	uniqueStories, err := findUniqueStoriesForUser(ctx)
	if err == nil {
		storyList, found := uniqueStories[sharedStory.SourceAlbumReference()]
		if found && len(storyList) > 1 {
			ctx.RequestLogger.Debug().Msg("multiple unique stories in list")
			isDuplicate = true
		} else if found {
			ctx.RequestLogger.Debug().Int("stories", len(storyList)).Msg("unique story list")
			for _, existingStory := range storyList {
				ctx.RequestLogger.Debug().Str("ref", sharedStory.SourceAlbumReference()).Str("existing ref", existingStory.SourceAlbumReference()).Msg("checking existing story")
				isDuplicate = isDuplicate || sharedStory.StoryID != existingStory.StoryID
			}
		} else {

		}
	} else {
		ctx.RequestLogger.Info().Err(err).Msg("error finding unique")
		return "", false
	}
	return duplicateStory, isDuplicate
}

// Save updates the db with the current StoryGroup state
func (storyGroup *StoryGroup) Save(ctx awsproxy.FTContext) error {
	ctx.RequestLogger.Debug().Str("storyId", storyGroup.StoryID).Str("groupId", storyGroup.GroupID).Msg("Putting story group")
	condtionExp := " attribute_not_exists(resourceId) or status=:statusRemoved"
	var expressionValues map[string]types.AttributeValue = make(map[string]types.AttributeValue)
	expressionValues[":statusRemoved"] = &types.AttributeValueMemberS{Value: "r"}
	if len(storyGroup.BaseVersion) > 0 {
		condtionExp = condtionExp + " or baseVersion=:baseVersion"
		expressionValues[":baseVersion"] = &types.AttributeValueMemberS{Value: storyGroup.BaseVersion}
	}
	storyGroup.BaseVersion = storyGroup.Version
	storyGroupMap, marshalError := attributevalue.MarshalMap(storyGroup)
	if nil != marshalError {
		return marshalError
	}
	resourceID := ftdb.ResourceIDFromGroupID(storyGroup.GroupID)
	storyGroupMap[ftdb.ResourceIDField] = &types.AttributeValueMemberS{Value: resourceID}
	referenceID := ftdb.ReferenceIDFromStoryID(storyGroup.StoryID)
	storyGroupMap[ftdb.ReferenceIDField] = &types.AttributeValueMemberS{Value: referenceID}
	_, err := ctx.DBSvc.PutItem(ctx.Context, &dynamodb.PutItemInput{
		TableName: aws.String(ftdb.GetTableName()),
		Item:      storyGroupMap,
	})
	if nil != err {
		ctx.RequestLogger.Error().Msg(fmt.Sprintf("save story group failed %s", err.Error()))
	} else {
		ctx.RequestLogger.Info().Msg("save story group succeeded")
	}
	return err
}

func (e StoryNotFoundError) Error() string {
	return fmt.Sprintf("No story with ID %s", e.StoryID)
}
