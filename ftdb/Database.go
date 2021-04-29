package ftdb

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
)

// FolktellsRecord a single record that combines the useful properties of
// all record types in the DB
type FolktellsRecord struct {
	ResourceID       string `json:"resourceId,omitempty" dynamodbav:"resourceId,omitempty"`
	ReferenceID      string `json:"referenceId,omitempty" dynamodbav:",omitempty"`
	Email            string `json:"email,omitempty" dynamodbav:"email,omitempty"`
	ID               string `json:"id,omitempty" dynamodbav:"id,omitempty"`
	Name             string `json:"name,omitempty" dynamodbav:"name,omitempty"`
	CallChannel      string `json:"callChannel,omitempty" dynamodbav:"callChannel,omitempty"`
	CallPeerId       string `json:"callPeerId,omitempty" dynamodbav:"callPeerId,omitempty"`
	InviteAccepted   string `json:"inviteAccepted,omitempty" dynamodbav:"inviteAccepted,omitempty"`
	SharingExpiry    int    `json:"sharingExpiry,omitempty" dynamodbav:"sharingExpiry,omitempty"`
	SharingProductID string `json:"sharingProductId,omitempty" dynamodbav:"sharingProductId,omitempty"`
	StorySource      string `json:"storySource,omitempty" dynamodbav:"storySource,omitempty"`
	AlbumReference   string `json:"albumReference,omitempty" dynamodbav:"albumReference,omitempty"`
	GroupID          string `json:"groupId,omitempty" dynamodbav:"groupId,omitempty"`
	MemberEmail      string `json:"memberEmail,omitempty" dynamodbav:"memberEmail,omitempty"`
	MemberID         string `json:"memberId,omitempty" dynamodbav:"memberId,omitempty"`
	CreatedAt        int    `json:"createdAt,omitempty" dynamodbav:"createdAt,omitempty"`
	LastUpdated      int    `json:"lastUpdated,omitempty" dynamodbav:"lastUpdated,omitempty"`
	LastUpdatedBy    string `json:"lastUpdatedBy,omitempty" dynamodbav:"lastUpdatedBy,omitempty"`
	InvitationID     string `json:"invitationId,omitempty" dynamodbav:"invitationId,omitempty"`
	InvitedByEmail   string `json:"invitedByEmail,omitempty" dynamodbav:"invitedByEmail,omitempty"`
	InvitedByID      string `json:"invitedById,omitempty" dynamodbav:"invitedById,omitempty"`
	AuthToken        string `json:"authToken,omitempty" dynamodbav:"authToken,omitempty"`
	ExternalID       string `json:"externalId,omitempty" dynamodbav:"externalId,omitempty"`
	P2PID            int    `json:"p2pId,omitempty" dynamodbav:"p2pId,omitempty"`
	Version          string `json:"version,omitempty" dynamodbav:"version,omitempty"`
	BaseVersion      string `json:"baseVersion,omitempty" dynamodbav:"baseVersion,omitempty"`
}

// GetItem load a single item from DynamoDB and unmarshall it into the given result
func GetItem(ftCtx awsproxy.FTContext, resourceID string, referenceID string, result interface{}) (bool, error) {
	dbResult, err := ftCtx.DBSvc.GetItem(ftCtx.Context, &dynamodb.GetItemInput{
		TableName: aws.String(GetTableName()),
		Key: map[string]types.AttributeValue{
			ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
			ReferenceIDField: &types.AttributeValueMemberS{Value: referenceID},
		},
	})
	if nil != err {
		return false, err
	}
	if nil == dbResult.Item {
		return false, nil
	}
	err = attributevalue.UnmarshalMap(dbResult.Item, result)
	if nil != err {
		return false, err
	}
	return true, nil

}

// DeleteItem remove a single item from DynamoDB
func DeleteItem(ftCtx awsproxy.FTContext, resourceID string, referenceID string) error {
	_, err := ftCtx.DBSvc.DeleteItem(ftCtx.Context, &dynamodb.DeleteItemInput{
		TableName: aws.String(GetTableName()),
		Key: map[string]types.AttributeValue{
			ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
			ReferenceIDField: &types.AttributeValueMemberS{Value: referenceID},
		},
	})
	return err

}

// PutItem adds a single item to DynamoDB
func PutItem(ftCtx awsproxy.FTContext, resourceID string, referenceID string, item interface{}) error {
	itemMap, marshalError := attributevalue.MarshalMap(item)
	if nil != marshalError {
		return marshalError
	}
	itemMap[ResourceIDField] = &types.AttributeValueMemberS{Value: resourceID}
	itemMap[ReferenceIDField] = &types.AttributeValueMemberS{Value: referenceID}
	_, err := ftCtx.DBSvc.PutItem(ftCtx.Context, &dynamodb.PutItemInput{
		TableName: aws.String(GetTableName()),
		Item:      itemMap,
	})
	return err
}

// UpdateItem updates a single item in DynamoDB
func UpdateItem(ftCtx awsproxy.FTContext, resourceID, referenceID, updateExpression string, updates interface{}) error {
	updateMap, marshalError := attributevalue.MarshalMap(updates)
	if nil != marshalError {
		return marshalError
	}
	_, err := ftCtx.DBSvc.UpdateItem(ftCtx.Context, &dynamodb.UpdateItemInput{
		TableName: aws.String(GetTableName()),
		Key: map[string]types.AttributeValue{
			ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
			ReferenceIDField: &types.AttributeValueMemberS{Value: referenceID},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeValues: updateMap,
	})
	return err
}

// Query search the db on a particular index
func Query(ftCtx awsproxy.FTContext, dbIndex, keyCondition string, keyAttributes map[string]types.AttributeValue) ([]FolktellsRecord, error) {
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:                 aws.String(GetTableName()),
		IndexName:                 aws.String(dbIndex),
		KeyConditionExpression:    aws.String(keyCondition),
		ExpressionAttributeValues: keyAttributes,
	})
	queryResults := []FolktellsRecord{}
	if err != nil {
		ftCtx.RequestLogger.Debug().Err(err).Msg("Query error")
		return queryResults, err
	} else if result.Count == 0 {
		ftCtx.RequestLogger.Debug().Msg("Query no matching requests")
		return queryResults, nil
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &queryResults)
	return queryResults, err
}

// QueryByResource find all records that match on the resourceID
func QueryByResource(ftCtx awsproxy.FTContext, resID string) ([]FolktellsRecord, error) {
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(GetTableName()),
		KeyConditionExpression: aws.String("resourceId = :rid1 "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rid1": &types.AttributeValueMemberS{Value: resID},
		},
	})
	queryResults := []FolktellsRecord{}
	if err != nil {
		ftCtx.RequestLogger.Debug().Err(err).Msg("Query error")
		return queryResults, err
	} else if result.Count == 0 {
		ftCtx.RequestLogger.Debug().Msg("Query no matching requests")
		return queryResults, nil
	}

	err = attributevalue.UnmarshalListOfMaps(result.Items, &queryResults)
	return queryResults, err
}

// returns the time in ms since the epoch
func NowMillisecondsSinceEpoch() int64 {
	now := time.Now()
	nanos := now.UnixNano()
	return nanos / 1000000
}
