package store

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// UserSubscription holds information about a subscription for a user and a product
type UserSubscription struct {
	StoreType             string
	UserID                string
	ProductID             string
	OriginalTransactionID string
	TransactionID         string
	ExpiresDateTimeMS     int64
	VerificationData      string
	AutoRenew             bool
	GracePeriodExpiresMS  int64
}

// UserSubscriptionNotFoundError returned by LoadUserSubscriptionByTransaction when the transaction is not in the DB
type UserSubscriptionNotFoundError struct {
	error
	TransactionID string
}

// UserSubscriptionMismatchError returned by LoadUserSubscriptionByTransaction when the transaction is not in the DB
type UserSubscriptionMismatchError struct {
	error
	ExistingUserID  string
	AttemptedUserID string
}

// LoadUserSubscriptionByTransaction returns either an error or UserSubscription
func LoadUserSubscriptionByTransaction(ctx awsproxy.FTContext, transactionID string) (*UserSubscription, error) {
	result, err := ctx.DBSvc.Query(ctx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		KeyConditionExpression: aws.String("resourceId = :subscriptionId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":subscriptionId": &types.AttributeValueMemberS{Value: ftdb.ResourceIDFromTransactionID(transactionID)},
		},
	})
	if err != nil {
		return nil, err
	} else if result.Count == 0 {
		return nil, &UserSubscriptionNotFoundError{TransactionID: transactionID}
	} else if result.Count > 1 {
		return nil, fmt.Errorf("Multiple subscriptions found for %s", transactionID)
	}
	subscriptionResults := []UserSubscription{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &subscriptionResults)
	if nil != err {
		return nil, err
	}
	return &subscriptionResults[0], nil
}

// Save updates the DB with response creating or updating the receipt record and
// the user record.
// The receipt record is an T#[TransactionID]. The user record will not be updated if
// the receipt is already associated with a different user.
func (userSubscription *UserSubscription) Save(ctx awsproxy.FTContext) error {
	existingSubscription, err := LoadUserSubscriptionByTransaction(ctx, userSubscription.OriginalTransactionID)
	newSubscription := false
	if nil != err {
		switch err.(type) {
		case *UserSubscriptionNotFoundError:
			newSubscription = true
		default:
			return err
		}
	}
	if !newSubscription && existingSubscription.UserID != userSubscription.UserID {
		return &UserSubscriptionMismatchError{ExistingUserID: existingSubscription.UserID,
			AttemptedUserID: userSubscription.UserID}
	}

	return userSubscription.saveToDB(ctx)
}

func (userSubscription *UserSubscription) saveToDB(ctx awsproxy.FTContext) error {
	resourceID := ftdb.ResourceIDFromTransactionID(userSubscription.OriginalTransactionID)
	referenceID := ftdb.ReferenceIDFromUserID(userSubscription.UserID)
	err := ftdb.PutItem(ctx, resourceID, referenceID, userSubscription)
	if nil != err {
		ctx.RequestLogger.Error().Str("transaction_id", userSubscription.OriginalTransactionID).Msg("save subscription failed")
	} else {
		ctx.RequestLogger.Info().Msg("save subscription succeeded")
	}
	return err
}

func (e UserSubscriptionNotFoundError) Error() string {
	return fmt.Sprintf("No subscription with ID %s", e.TransactionID)
}

func (e UserSubscriptionMismatchError) Error() string {
	return fmt.Sprintf("Subscription for %s tried to udpate with %s", e.ExistingUserID, e.AttemptedUserID)
}
