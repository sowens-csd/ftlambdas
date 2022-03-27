package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	uuid "github.com/satori/go.uuid"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
	"github.com/sowens-csd/ftlambdas/sharing"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	return createGroup(ftCtx, request.Body), nil
}

func createGroup(ftCtx awsproxy.FTContext, groupJSON string) awsproxy.Response {
	shareGroupJSON := []byte(groupJSON)
	var shareGroup sharing.ShareGroup
	err := json.Unmarshal(shareGroupJSON, &shareGroup)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger)
	}
	ftCtx.RequestLogger.Info().Msg(fmt.Sprintf("Name after JSON parsing is %s", shareGroup.Name))
	email, userErr := findOwnerEmail(ftCtx, ftCtx.UserID)
	if nil != userErr {
		return awsproxy.HandleError(userErr, ftCtx.RequestLogger)
	}
	_, dbErr := insertGroupIntoDb(ftCtx, shareGroup)
	if nil != dbErr {
		return awsproxy.HandleError(dbErr, ftCtx.RequestLogger)
	}
	ownErr := insertOwnerInvitationIntoDb(ftCtx, shareGroup, email)
	if nil != ownErr {
		return awsproxy.HandleError(ownErr, ftCtx.RequestLogger)
	}
	return awsproxy.NewSuccessResponse(ftCtx)
}

func findOwnerEmail(ftCtx awsproxy.FTContext, userID string) (string, error) {
	resID := ftdb.ResourceIDFromUserID(userID)
	result, err := ftCtx.DBSvc.GetItem(ftCtx.Context, &dynamodb.GetItemInput{
		TableName: aws.String(ftdb.GetTableName()),
		Key: map[string]types.AttributeValue{
			ftdb.ResourceIDField:  &types.AttributeValueMemberS{Value: resID},
			ftdb.ReferenceIDField: &types.AttributeValueMemberS{Value: resID},
		},
	})
	if nil != err {
		return "", err
	}
	return result.Item[ftdb.EmailField].(*types.AttributeValueMemberS).Value, nil
}

func insertGroupIntoDb(ftCtx awsproxy.FTContext, shareGroup sharing.ShareGroup) (*dynamodb.PutItemOutput, error) {
	resourceID := ftdb.ResourceIDFromGroupID(shareGroup.GroupID)
	referenceID := resourceID
	ftCtx.RequestLogger.Info().Msg(fmt.Sprintf("AWS Name is %s", *aws.String(shareGroup.Name)))
	result, err := ftCtx.DBSvc.PutItem(ftCtx.Context, &dynamodb.PutItemInput{
		TableName: aws.String(ftdb.GetTableName()),
		Item: map[string]types.AttributeValue{
			ftdb.ResourceIDField:   &types.AttributeValueMemberS{Value: resourceID},
			ftdb.ReferenceIDField:  &types.AttributeValueMemberS{Value: referenceID},
			ftdb.IDField:           &types.AttributeValueMemberS{Value: shareGroup.GroupID},
			ftdb.NameField:         &types.AttributeValueMemberS{Value: shareGroup.Name},
			ftdb.OwnerIDField:      &types.AttributeValueMemberS{Value: ftCtx.UserID},
			ftdb.InvitationIDField: &types.AttributeValueMemberS{Value: shareGroup.InvitationID},
		},
	})
	ftCtx.RequestLogger.Info().Msg("insert result")
	return result, err
}

func insertOwnerInvitationIntoDb(ftCtx awsproxy.FTContext, shareGroup sharing.ShareGroup, email string) error {
	resourceID := ftdb.ResourceIDFromGroupID(shareGroup.GroupID)
	referenceID := ftdb.ReferenceIDFromUserID(ftCtx.UserID)
	versionUUID := uuid.NewV4()
	version := versionUUID.String()
	ftCtx.RequestLogger.Info().Str("uuid", version).Msg("got UUID")
	now := time.Now()
	lastUpdated := int(now.UTC().Unix() * 1000)
	err := ftdb.PutItem(ftCtx, resourceID, referenceID, sharing.GroupMember{
		GroupID:        shareGroup.GroupID,
		InvitationID:   shareGroup.InvitationID,
		InvitedByID:    ftCtx.UserID,
		MemberEmail:    email,
		MemberID:       ftCtx.UserID,
		InviteAccepted: sharing.MembershipAccepted,
		InvitedOn:      lastUpdated,
		Version:        version,
		BaseVersion:    version,
		LastUpdated:    lastUpdated,
		LastUpdatedBy:  ftCtx.UserID,
	})
	if nil != err {
		ftCtx.RequestLogger.Error().Err(err).Msg("invite failed")
	} else {
		ftCtx.RequestLogger.Info().Msg("invite succeeded")
	}
	return err
}

func main() {
	lambda.Start(Handler)
}
