package sharing

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// GroupNotFoundError returned by Load when the group is not in the DB
type GroupNotFoundError struct {
	error
	GroupID string
}

// ShareGroup is the information exchanged about a single group
type ShareGroup struct {
	GroupID      string `json:"id" dynamodbav:"id"`
	Name         string `json:"name" dynamodbav:"name"`
	OwnerID      string `json:"ownerId" dynamodbav:"ownerId"`
	InvitationID string `json:"invitationId" dynamodbav:"invitationId"`
}

// ResourceID returns the ResourceID for this member
func (group *ShareGroup) ResourceID() string {
	return ftdb.ReferenceIDFromGroupID(group.GroupID)
}

// ReferenceID returns the ReferenceID for this member
func (group *ShareGroup) ReferenceID() string {
	return ftdb.ReferenceIDFromUserID(group.GroupID)
}

// LoadGroup returns the group information matching the groupID or a GroupNotFound error
func LoadGroup(ftCtx awsproxy.FTContext, groupID string) (*ShareGroup, error) {
	resourceID := ftdb.ResourceIDFromGroupID(groupID)
	shareGroup := ShareGroup{}
	found, err := ftdb.GetItem(ftCtx, resourceID, resourceID, &shareGroup)
	if nil != err {
		return nil, err
	}
	if !found {
		return nil, &GroupNotFoundError{GroupID: groupID}
	}
	return &shareGroup, nil
}

// Delete deletes the group and any members
func (group *ShareGroup) Delete(ftCtx awsproxy.FTContext) error {

	references, err := group.findReferences(ftCtx)
	if err != nil {
		return err
	}
	for _, reference := range references {
		delete(ftCtx, group.ResourceID(), reference)
	}
	return delete(ftCtx, group.ResourceID(), group.ReferenceID())
}

func (group *ShareGroup) findReferences(ftCtx awsproxy.FTContext) ([]string, error) {
	result, err := ftCtx.DBSvc.Query(ftCtx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		KeyConditionExpression: aws.String("resourceId = :groupID "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":groupID": &types.AttributeValueMemberS{Value: ftdb.ResourceIDFromGroupID(group.GroupID)},
		},
	})
	var references []string
	keyResults := []ftdb.SharingCompositeKey{}
	if err != nil {
		return references, err
	} else if result.Count == 0 {
		return references, nil
	}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &keyResults)
	if nil != err {
		return references, err
	}
	for _, reference := range keyResults {
		references = append(references, reference.ReferenceID)
	}
	return references, err
}

func delete(ftCtx awsproxy.FTContext, resourceID string, referenceID string) error {
	err := ftdb.DeleteItem(ftCtx, resourceID, referenceID)
	if nil != err {
		ftCtx.RequestLogger.Error().Str("ResourceID", resourceID).Str("ReferenceID", referenceID).Err(err).Msg("Delete failed")
	}
	return err
}
