package sharing

import (
	"crypto/rand"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// ManagedUser is one that doesn't have an email address, usually
// created by a provisioning service rather than directly by the user
type ManagedUser struct {
	ID                    string                    `json:"id" dynamodbav:"id"`
	Name                  string                    `json:"name" dynamodbav:"name"`
	CreatedAt             int                       `json:"createdAt" dynamodbav:"createdAt"`
	CreatedBy             string                    `json:"createdBy" dynamodbav:"createdBy"`
	InviteAccepted        string                    `json:"inviteAccepted" dynamodbav:"inviteAccepted"`
	SharingProductID      string                    `json:"sharingProductId,omitempty" dynamodbav:"sharingProductId"`
	SharingExpiry         int64                     `json:"sharingExpiry,omitempty" dynamodbav:"sharingExpiry"`
	AutoRenew             bool                      `json:"autoRenew,omitempty" dynamodbav:"autoRenew"`
	GracePeriod           int64                     `json:"gracePeriod,omitempty" dynamodbav:"gracePeriod"`
	OriginalTransactionID string                    `json:"originalTransactionId,omitempty" dynamodbav:"originalTransactionId"`
	LastExpiryVerify      int64                     `json:"lastExpiryVerify,omitempty" dynamodbav:"lastExpiryVerify"`
	Phone                 string                    `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	CallChannel           string                    `json:"callChannel,omitempty" dynamodbav:"callChannel,omitempty"`
	CallPeerId            string                    `json:"callPeerId,omitempty" dynamodbav:"callPeerId,omitempty"`
	DeviceTokens          []DeviceNotificationToken `json:"deviceTokens,omitempty" dynamodbav:"deviceTokens,omitempty"`
	Deleted               string                    `json:"deleted,omitempty" dynamodbav:"deleted,omitempty"`
}

const managedUserIDLen = 8

// NewManagedUser creates a new managed user that can automatically log in
// exactly once by entering just the username
func NewManagedUser(userID, name string) ManagedUser {
	return ManagedUser{
		ID:   userID,
		Name: name,
	}
}

// LoadManagedUser attempt to load a user given their userID, returns a
// UserNotFoundError if the user is not found.
func LoadManagedUser(ctx awsproxy.FTContext, userID string) (*ManagedUser, error) {
	ctx.RequestLogger.Debug().Str("userID", userID).Msg("load Managed user")
	resourceID := ftdb.ResourceIDFromManagedUserID(userID)
	managedUser := ManagedUser{}
	found, err := ftdb.GetItem(ctx, resourceID, resourceID, &managedUser)
	if nil != err {
		return nil, err
	}
	if !found {
		return nil, &UserNotFoundError{UserID: userID}
	}
	ctx.RequestLogger.Debug().Msg("loaded Managed user")
	return &managedUser, nil
}

// AddManagedUser creates an entry for a new managed user in the DB.
func AddManagedUser(ctx awsproxy.FTContext, username, createdBy string) (*ManagedUser, error) {
	tryingID := true
	var managedUser ManagedUser
	for tryingID {
		userID, err := generateManagedUserID()
		if nil != err {
			ctx.RequestLogger.Info().Err(err).Msg("Error generating managed userID")
			return nil, err
		}
		ctx.RequestLogger.Debug().Str("name", username).Str("userID", userID).Msg("Add managed user")
		resourceID := ftdb.ResourceIDFromManagedUserID(userID)
		managedUser = ManagedUser{
			ID:        userID,
			Name:      username,
			CreatedBy: createdBy,
			CreatedAt: int(now()),
		}
		err = ftdb.PutUniqueItem(ctx, resourceID, resourceID, managedUser)
		if nil != err {
			switch err.(type) {
			case *types.ConditionalCheckFailedException:
				continue
			default:
				ctx.RequestLogger.Info().Err(err).Msg("Error adding managed user")
				return nil, err
			}
		}
		tryingID = false
	}
	ctx.RequestLogger.Debug().Msg("Added managed user")
	return &managedUser, nil
}

const allowedUserIDChars = "1234567890abcdefghijklmnopqrstuvwxyz"

func generateManagedUserID() (string, error) {
	buffer := make([]byte, managedUserIDLen)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(allowedUserIDChars)
	for i := 0; i < managedUserIDLen; i++ {
		buffer[i] = allowedUserIDChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}
