package sharing

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ReneKroon/ttlcache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	uuid "github.com/satori/go.uuid"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// UserInviteAccepted is the value for InviteAccepted when the user has
// signed up in Cognito.
const UserInviteAccepted = "y"

// UserInvitePending is the value for InviteAccepted when the invitation has
// been extended but the user has not yet signed up.
const UserInvitePending = "p"

const oneHour = 1 * 1000 * 60 * 60

const oneDay = 1 * 1000 * 60 * 60 * 24
const oneWeek = 1 * 1000 * 60 * 60 * 24 * 7

// DeviceNotificationToken the relationship between an app install on a particular
// device, the notification token for that device and the SNS Endpoint that can be
// used to notify the device.
type DeviceNotificationToken struct {
	AppInstallID      string `json:"deviceInstallId" dynamodbav:"deviceInstallId"`
	SNSEndpoint       string `json:"snsEndpoint" dynamodbav:"snsEndpoint"`
	NotificationToken string `json:"notificationToken" dynamodbav:"notificationToken"`
	AppVersion        string `json:"appVersion" dynamodbav:"appVersion"`
}

// OnlineUser is the basic profile information
type OnlineUser struct {
	ID                    string                    `json:"id" dynamodbav:"id"`
	Name                  string                    `json:"name" dynamodbav:"name"`
	Email                 string                    `json:"email" dynamodbav:"email"`
	CreatedAt             string                    `json:"createdAt" dynamodbav:"createdAt"`
	InviteAccepted        string                    `json:"inviteAccepted" dynamodbav:"inviteAccepted"`
	SharingProductID      string                    `json:"sharingProductId,omitempty" dynamodbav:"sharingProductId"`
	SharingExpiry         int64                     `json:"sharingExpiry,omitempty" dynamodbav:"sharingExpiry"`
	AutoRenew             bool                      `json:"autoRenew,omitempty" dynamodbav:"autoRenew"`
	GracePeriod           int64                     `json:"gracePeriod,omitempty" dynamodbav:"gracePeriod"`
	OriginalTransactionID string                    `json:"originalTransactionId,omitempty" dynamodbav:"originalTransactionId"`
	LastExpiryVerify      int64                     `json:"lastExpiryVerify,omitempty" dynamodbav:"lastExpiryVerify"`
	EmailOptOut           bool                      `json:"emailOptOut" dynamodbav:"emailOptOut"`
	EmailOptOutReason     string                    `json:"emailOptOutReason,omitempty" dynamodbav:"emailOptOutReason"`
	Phone                 string                    `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	CallChannel           string                    `json:"callChannel,omitempty" dynamodbav:"callChannel,omitempty"`
	CallPeerId            string                    `json:"callPeerId,omitempty" dynamodbav:"callPeerId,omitempty"`
	TwilioPhoneSID        string                    `json:"twilioPhoneSid,omitempty" dynamodbav:"twilioPhoneSid,omitempty"`
	DeviceTokens          []DeviceNotificationToken `json:"deviceTokens,omitempty" dynamodbav:"deviceTokens,omitempty"`
	Deleted               string                    `json:"deleted,omitempty" dynamodbav:"deleted,omitempty"`
}

type mailIndexQueryResult struct {
	ID          string `json:"id" dynamodbav:"id"`
	Email       string `json:"email" dynamodbav:"email"`
	ResourceID  string `json:"resourceId" dynamodbav:"resourceId"`
	ReferenceID string `json:"referenceId" dynamodbav:"referenceId"`
}

type phoneIndexQueryResult struct {
	ID    string `json:"id" dynamodbav:"id"`
	Phone string `json:"phone" dynamodbav:"phone"`
}

// UserNotFoundError returned by LoadByEmail or Load when the user is not in the DB
type UserNotFoundError struct {
	error
	Email  string
	UserID string
	Phone  string
}

// UserAlreadyExistsError returned by Save when the user is already in the DB
type UserAlreadyExistsError struct {
	error
	Email string
}

var userByPhoneCache ttlcache.Cache

func init() {
	userByPhoneCache := ttlcache.NewCache()
	userByPhoneCache.SetTTL(time.Duration(1 * time.Hour))
	userByPhoneCache.SkipTtlExtensionOnHit(true)
}

// NewOnlineUser creates an accepted user, i.e. one that has created an account in the
// authentication identity provider (currently Cognito)
func NewOnlineUser(userID string, name string, email string) OnlineUser {
	return OnlineUser{
		ID:             userID,
		Name:           name,
		Email:          email,
		InviteAccepted: UserInviteAccepted,
	}
}

// NewOnlineUserInvitation creates an invitation for a pending user, i.e. one that has created
// been invited to join Folktells by another user but has not yet created  an account
// in the authentication identity provider (currently Cognito)
func NewOnlineUserInvitation(userID string, name string, email string) OnlineUser {
	return OnlineUser{
		ID:             userID,
		Name:           name,
		Email:          email,
		InviteAccepted: UserInvitePending,
	}
}

// NewTemporaryUserID creates a value usable in the UserID field of an OnlineUser.
func NewTemporaryUserID() (string, error) {
	uid := uuid.NewV4()
	return uid.String(), nil
}

// UpdateOnlineUser ensures that only allowed fields of the user are properly updated given
// the JSON that is provided.
func UpdateOnlineUser(ctx awsproxy.FTContext, onlineUserJSON string) error {
	inputJSON := []byte(onlineUserJSON)
	var onlineUser OnlineUser
	err := json.Unmarshal(inputJSON, &onlineUser)
	if nil != err {
		return err
	}
	err = onlineUser.Update(ctx)
	if nil != err {
		return err
	}
	return err
}

// UpdateAll ensures that all non-destructive  fields of the user are properly updated
func (ou *OnlineUser) UpdateAll(ctx awsproxy.FTContext) error {
	return ou.saveUserToDB(ctx)
}

// IsAccepted returns true if the user has an account in the authentication provider
func (ou *OnlineUser) IsAccepted() bool {
	return UserInviteAccepted == ou.InviteAccepted
}

// IsPending returns true if the user has not yet created an account in the
// authentication provider
func (ou *OnlineUser) IsPending() bool {
	return !ou.IsAccepted()
}

// AllowsEmail returns true unless the user has explicitly opted out
func (ou *OnlineUser) AllowsEmail() bool {
	return !ou.EmailOptOut
}

// HasPhone returns true if the user has a phone number attached to their account
func (ou *OnlineUser) HasPhone() bool {
	return len(ou.Phone) > 0
}

// IsSubscriptionCheckRequired returns true if the user has a subscription nearing expiry
// that has not been checked recently
func (ou *OnlineUser) IsSubscriptionCheckRequired() bool {
	if ou.SharingExpiry == 0 {
		return false
	}
	now := now()
	timeToExpiry := now - ou.SharingExpiry
	timeSinceLastVerify := now - ou.LastExpiryVerify
	timeBetweenVerify := oneWeek
	if timeToExpiry < oneDay {
		timeBetweenVerify = oneHour
	}
	return timeSinceLastVerify > int64(timeBetweenVerify)
}

// IsSubscriptionExpired returns true if the user does not have an active subscription
func (ou *OnlineUser) IsSubscriptionExpired() bool {
	if ou.SharingExpiry == 0 {
		return true
	}
	now := now()
	timeToExpiry := ou.SharingExpiry - now
	if ou.GracePeriod != 0 {
		timeToExpiry = ou.GracePeriod - now
	}
	return timeToExpiry <= 0
}

// FindDeviceNotificationToken returns the notification token for the given appInstallID if one
// exists, otherwise returns nil
func (ou *OnlineUser) FindDeviceNotificationToken(appInstallID string) *DeviceNotificationToken {
	index := ou.findNotificationTokenIndex(appInstallID)
	if -1 == index {
		return nil
	}
	deviceToken := ou.DeviceTokens[index]
	return &DeviceNotificationToken{
		AppInstallID:      deviceToken.AppInstallID,
		NotificationToken: deviceToken.NotificationToken,
		SNSEndpoint:       deviceToken.SNSEndpoint}
}

// SetDeviceNotificationToken either adds a new entry with the endpoint and token or updates an existing entry if
// the appInstallID already has an endpoint for the user.
func (ou *OnlineUser) SetDeviceNotificationToken(appInstallID, endpoint, notificationToken, appVersion string) {
	index := ou.findNotificationTokenIndex(appInstallID)
	if index >= 0 {
		ou.DeviceTokens[index].SNSEndpoint = endpoint
		ou.DeviceTokens[index].NotificationToken = notificationToken
		ou.DeviceTokens[index].AppVersion = appVersion
		return
	}
	if nil == ou.DeviceTokens {
		ou.DeviceTokens = make([]DeviceNotificationToken, 0)
	}
	ou.DeviceTokens = append(ou.DeviceTokens, DeviceNotificationToken{
		AppInstallID:      appInstallID,
		NotificationToken: notificationToken,
		SNSEndpoint:       endpoint,
		AppVersion:        appVersion})
}

func (ou *OnlineUser) findNotificationTokenIndex(appInstallID string) int {
	if nil == ou.DeviceTokens {
		return -1
	}
	for idx, deviceToken := range ou.DeviceTokens {
		if appInstallID == deviceToken.AppInstallID {
			return idx
		}
	}
	return -1
}

// Save updates the db with the current user state
func (ou *OnlineUser) Save(ctx awsproxy.FTContext) error {
	existingOU, err := LoadOnlineUserByEmail(ctx, ou.Email)
	if nil != err {
		switch err.(type) {
		case *UserNotFoundError:
			break
		default:
			return err
		}
	}
	updateMembers := false
	if nil != existingOU {
		if ou.IsPending() || existingOU.IsAccepted() {
			return &UserAlreadyExistsError{Email: ou.Email}
		}
		updateMembers = true
		err := existingOU.Delete(ctx)
		if nil != err {
			return err
		}
	}
	err = ou.saveUserToDB(ctx)
	if !updateMembers || nil != err {
		return err
	}
	return SwitchAllMembersToPermanentUser(ctx, ou.ID, existingOU.ID)
}

// LoadOrCreateTemporary either loads an existing user with that email address or if
// there isn't one creates a new temporary user, saves that, and returns it.
//
// The returned user can be used as though it were a permanent user, though if you
// need to differentiate use the IsPending() test on the returned user to check.
func LoadOrCreateTemporary(ctx awsproxy.FTContext, name string, email string) (*OnlineUser, error) {
	existingOU, err := LoadOnlineUserByEmail(ctx, email)
	if nil != err {
		switch err.(type) {
		case *UserNotFoundError:
			userID, uidErr := NewTemporaryUserID()
			if nil != uidErr {
				return nil, uidErr
			}
			newUser := NewOnlineUserInvitation(userID, email, email)
			saveErr := newUser.Save(ctx)
			if nil != saveErr {
				return nil, saveErr
			}
			return &newUser, nil
			break
		default:
			return nil, err
		}
	}
	return existingOU, nil
}

// LoadOnlineUser attempt to load a user given their userID, returns a
// UserNotFoundError if the user is not found.
func LoadOnlineUser(ctx awsproxy.FTContext, userID string) (*OnlineUser, error) {
	ctx.RequestLogger.Debug().Str("userID", userID).Msg("load user")
	resourceID := ftdb.ResourceIDFromUserID(userID)
	onlineUser := OnlineUser{}
	found, err := ftdb.GetItem(ctx, resourceID, resourceID, &onlineUser)
	if nil != err {
		return nil, err
	}
	if !found {
		return nil, &UserNotFoundError{UserID: userID}
	}
	ctx.RequestLogger.Debug().Msg("loaded user")
	return &onlineUser, nil
}

// LoadOnlineUserByEmail attempts to load a user given their email, returns a
// UserNotFoundError if the user is not found.
func LoadOnlineUserByEmail(ctx awsproxy.FTContext, email string) (*OnlineUser, error) {
	result, err := ctx.DBSvc.Query(ctx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		IndexName:              aws.String(ftdb.EmailIndex),
		KeyConditionExpression: aws.String("email = :email "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: strings.ToLower(email)},
		},
	})
	if err != nil {
		return nil, err
	} else if result.Count == 0 {
		ctx.RequestLogger.Debug().Str("email", email).Msg("no user with email")
		return nil, &UserNotFoundError{Email: email}
	}

	byMailResults := []mailIndexQueryResult{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &byMailResults)
	if nil != err {
		return nil, err
	}
	var userID string
	for _, mailResult := range byMailResults {
		if ftdb.IsUserResource(mailResult.ResourceID) && ftdb.IsUserReference(mailResult.ReferenceID) {
			if len(userID) > 0 {
				return nil, fmt.Errorf("Multiple users found for %s", email)
			}
			userID = mailResult.ID
		}
	}

	return LoadOnlineUser(ctx, userID)
}

// LoadOnlineUserByPhone attempts to load a user given the phone number associated
// with the account, returns a UserNotFoundError if the user is not found.
func LoadOnlineUserByPhone(ctx awsproxy.FTContext, phone string) (*OnlineUser, error) {
	if value, exists := userByPhoneCache.Get(phone); exists {
		return value.(*OnlineUser), nil
	}
	result, err := ctx.DBSvc.Query(ctx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		IndexName:              aws.String(ftdb.PhoneIndex),
		KeyConditionExpression: aws.String("phone = :phone "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":phone": &types.AttributeValueMemberS{Value: strings.ToLower(phone)},
		},
	})
	if err != nil {
		return nil, err
	} else if result.Count == 0 {
		return nil, &UserNotFoundError{Phone: phone}
	} else if result.Count > 1 {
		return nil, fmt.Errorf("Multiple users found for %s", phone)
	}
	byPhoneResults := []phoneIndexQueryResult{}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &byPhoneResults)
	if nil != err {
		return nil, err
	}

	onlineUser, err := LoadOnlineUser(ctx, byPhoneResults[0].ID)
	if nil != err {
		userByPhoneCache.Set(phone, onlineUser)
	}
	return onlineUser, err
}

// Delete from the DB by their UserID
func (ou *OnlineUser) Delete(ctx awsproxy.FTContext) error {
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	return ftdb.DeleteItem(ctx, resourceID, resourceID)
}

// UpdateSharingSubscription make sure the subscription expiry for the given product
// matches the provided expiry and is persisted.
func (ou *OnlineUser) UpdateSharingSubscription(ctx awsproxy.FTContext, productID string, expiry int64, autoRenew bool, gracePeriod int64, originalTransactionID string) error {
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	type sharingSubscriptionUpdate struct {
		ProductID             string `json:":p" dynamodbav:":p"`
		SharingExpiry         int64  `json:":s" dynamodbav:":s"`
		GracePeriod           int64  `json:":g" dynamodbav:":g"`
		AutoRenew             bool   `json:":a" dynamodbav:":a"`
		LastExpiryVerify      int64  `json:":l" dynamodbav:":l"`
		OriginalTransactionID string `json:":o" dynamodbav:":o"`
	}
	now := now()
	update, err := attributevalue.MarshalMap(sharingSubscriptionUpdate{
		ProductID:             productID,
		SharingExpiry:         expiry,
		GracePeriod:           gracePeriod,
		AutoRenew:             autoRenew,
		LastExpiryVerify:      now,
		OriginalTransactionID: originalTransactionID,
	})
	if err != nil {
		return err
	}
	_, err = ctx.DBSvc.UpdateItem(ctx.Context, &dynamodb.UpdateItemInput{
		TableName: aws.String(ftdb.GetTableName()),
		Key: map[string]types.AttributeValue{
			ftdb.ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
			ftdb.ReferenceIDField: &types.AttributeValueMemberS{Value: resourceID},
		},
		UpdateExpression:          aws.String("set sharingProductId = :p, sharingExpiry = :s, gracePeriod = :g, autoRenew = :a, lastExpiryVerify= :l, originalTransactionId = :o"),
		ExpressionAttributeValues: update,
	})
	if err == nil {
		ou.SharingProductID = productID
		ou.SharingExpiry = expiry
		ou.LastExpiryVerify = now
	} else {
		ctx.RequestLogger.Info().Msg(fmt.Sprintf("Failed to save sharing subscription %s", err.Error()))
	}
	return err
}

func (ou *OnlineUser) saveUserToDB(ctx awsproxy.FTContext) error {
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	referenceID := resourceID
	now := time.Now()
	sec := int(now.UTC().Unix() * 1000)
	if len(ou.CreatedAt) == 0 {
		ou.CreatedAt = strconv.Itoa(sec)
	}
	err := ftdb.PutItem(ctx, resourceID, referenceID, ou)
	if nil != err {
		ctx.RequestLogger.Error().Str("userID", ou.ID).Msg("createUser failed")
	} else {
		ctx.RequestLogger.Info().Msg("createUser succeeded")
	}
	return err
}

// AcceptInvitation change the accepted flag to y
func (ou *OnlineUser) AcceptInvitation(ctx awsproxy.FTContext) error {
	ctx.RequestLogger.Debug().Str("userID", ou.ID).Msg("acceptInvite")
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	type userUpdate struct {
		InviteAccepted string `json:":i" dynamodbav:":i"`
	}
	err := ftdb.UpdateItem(ctx, resourceID, resourceID, "set inviteAccepted = :i", userUpdate{
		InviteAccepted: UserInviteAccepted,
	})
	if nil != err {
		ctx.RequestLogger.Error().Str("userID", ou.ID).Msg("accept invite failed")
	} else {
		ctx.RequestLogger.Info().Str("userID", ou.ID).Msg("accept invite succeeded")
	}
	return err
}

// Update modifies only allowed fields in the DB
func (ou *OnlineUser) Update(ctx awsproxy.FTContext) error {
	ctx.RequestLogger.Debug().Str("userID", ou.ID).Str("name", ou.Name).Bool("OptOut", ou.EmailOptOut).Msg("Updating to")
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	type userUpdate struct {
		Name        string `json:":d" dynamodbav:":d"`
		EmailOptOut bool   `json:":e" dynamodbav:":e"`
		CallChannel string `json:":c" dynamodbav:":c"`
		CallPeerId  string `json:":p" dynamodbav:":p"`
	}
	update, err := attributevalue.MarshalMap(userUpdate{
		Name:        ou.Name,
		EmailOptOut: ou.EmailOptOut,
		CallChannel: ou.CallChannel,
		CallPeerId:  ou.CallPeerId,
	})
	if err != nil {
		return err
	}
	_, err = ctx.DBSvc.UpdateItem(ctx.Context, &dynamodb.UpdateItemInput{
		TableName: aws.String(ftdb.GetTableName()),
		Key: map[string]types.AttributeValue{
			ftdb.ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
			ftdb.ReferenceIDField: &types.AttributeValueMemberS{Value: resourceID},
		},
		UpdateExpression:          aws.String("set name = :d, emailOptOut = :e, channelName = :c, callPeerId = :p"),
		ExpressionAttributeValues: update,
	})

	if nil != err {
		ctx.RequestLogger.Error().Err(err).Str("userID", ou.ID).Msg("update user failed")
	} else {
		ctx.RequestLogger.Info().Msg("Update user succeeded")
	}
	return err
}

// Update modifies only allowed fields in the DB
func (ou *OnlineUser) UpdateDeviceTokens(ctx awsproxy.FTContext) error {
	ctx.RequestLogger.Debug().Msg("updating device tokens")
	resourceID := ftdb.ResourceIDFromUserID(ou.ID)
	type deviceTokenUpdate struct {
		DeviceTokens []DeviceNotificationToken `json:":d,omitempty" dynamodbav:":d,omitempty"`
	}
	err := ftdb.UpdateItem(ctx, resourceID, resourceID, "set deviceTokens = :d", deviceTokenUpdate{
		DeviceTokens: ou.DeviceTokens,
	})
	if nil != err {
		ctx.RequestLogger.Error().Err(err).Str("userID", ou.ID).Msg("update device tokens failed")
	} else {
		ctx.RequestLogger.Debug().Msg("update device tokens succeeded")
	}
	return err
}

// returns the time in ms since the epoch
func now() int64 {
	now := time.Now()
	nanos := now.UnixNano()
	return nanos / 1000000
}

func (e UserNotFoundError) Error() string {
	if len(e.Email) > 0 {
		return fmt.Sprintf("No user with email %s", e.Email)
	} else {
		return fmt.Sprintf("No user with ID %s", e.UserID)
	}
}

func (e UserAlreadyExistsError) Error() string {
	return fmt.Sprintf("Already a user for this email %s", e.Email)
}
