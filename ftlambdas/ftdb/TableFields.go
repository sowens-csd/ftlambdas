package ftdb

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// EndpointHostResourceID the ResourceID and ReferenceID of the record that contains the endpoint host
// name for web socket use
const EndpointHostResourceID = "CFG#EndpointHost"

// All group records have this prefix, group to group, group to member and group to story
const groupKeyPrefix = "G#"

// All user records have this prefix, currently only user to user
const userKeyPrefix = "U#"

// All story records have this prefix, currently only story to story
const storyKeyPrefix = "S#"

// A transaction is a receipt for purchase from one of the app stores
const transactionKeyPrefix = "T#"

// This prefix identifies emails that are used to authenticate via the P2P provider, Connectycube
const p2pKeyPrefix = "P#"

// An authentication request is an uneverified request to authenticate a particular email address
const authRequestKeyPrefix = "AR#"

// An authentication token is a verified request for a particular user
const authTokenKeyPrefix = "AT#"

// DeleteRemove means all records related to this should be removed permanently from the DB
const DeleteRemove = "Remove"

// DeleteUse means this record should no longer be used
const DeleteUse = "Use"

// ResourceIDForAuthRequest creates a hash key from the date
func ResourceIDForAuthRequest() string {
	now := time.Now()
	year, month, day := now.Date()
	hour := now.Hour()
	return fmt.Sprintf("%s%d-%d-%d:%d", authRequestKeyPrefix, year, month, day, hour)
}

// ResourceIDForPrevAuthRequest creates a hash key from the previous hour's date
func ResourceIDForPrevAuthRequest(hours int) string {
	lastHour := time.Now().Add(time.Hour * time.Duration(0-hours))
	year, month, day := lastHour.Date()
	hour := lastHour.Hour()
	return fmt.Sprintf("%s%d-%d-%d:%d", authRequestKeyPrefix, year, month, day, hour)
}

// ReferenceIDFromAuthRequestID creates a hash key from an
// authorization request ID
func ReferenceIDFromAuthRequestID(requestID string) string {
	return authRequestKeyPrefix + requestID
}

// ResourceIDFromEmail creates a hash key from an encrypted email
func ResourceIDFromEmail(encryptedEmail string) string {
	return authTokenKeyPrefix + encryptedEmail
}

// ReferenceIDFromAuthTokenHash creates a partition key from an authorization token
func ReferenceIDFromAuthTokenHash(authTokenHash string) string {
	return authTokenKeyPrefix + authTokenHash
}

// AuthTokenHashFromReferenceID extracts the auth token hash from a hash key made from that token hash
func AuthTokenHashFromReferenceID(referenceID string) string {
	referenceRunes := []rune(referenceID)
	return string(referenceRunes[3:])
}

// ResourceIDFromGroupID creates a hash key from a groupID
func ResourceIDFromGroupID(groupID string) string {
	return groupKeyPrefix + groupID
}

// ReferenceIDFromGroupID creates a hash key from a groupID
func ReferenceIDFromGroupID(groupID string) string {
	return groupKeyPrefix + groupID
}

// ResourceIDFromUserID creates a hash key from a groupID
func ResourceIDFromUserID(userID string) string {
	return userKeyPrefix + userID
}

// IsUserReference returns true if the given referenceID (hash key) refers to a user
func IsUserReference(referenceID string) bool {
	return strings.HasPrefix(referenceID, userKeyPrefix)
}

// IsUserResource returns true if the given resourceID (hash key) refers to a user
func IsUserResource(resourceID string) bool {
	return strings.HasPrefix(resourceID, userKeyPrefix)
}

// IsAuthReference returns true if the given referenceID (hash key) refers to an authentication record
func IsAuthReference(referenceID string) bool {
	return strings.HasPrefix(referenceID, authTokenKeyPrefix)
}

// IsAuthResource returns true if the given resourceID (hash key) refers to an authentication record
func IsAuthResource(resourceID string) bool {
	return strings.HasPrefix(resourceID, authTokenKeyPrefix)
}

// IsP2PAuthResource returns true if the given referenceID (hash key) refers to a P2P authentication record
func IsP2PAuthResource(referenceID string) bool {
	return strings.HasPrefix(referenceID, p2pKeyPrefix)
}

// ReferenceIDFromUserID creates a hash key from a groupID
func ReferenceIDFromUserID(userID string) string {
	return userKeyPrefix + userID
}

// UserIDFromResourceID extracts the userID from a hash key made from that userID
func UserIDFromResourceID(userResourceID string) string {
	userRunes := []rune(userResourceID)
	return string(userRunes[2:])
}

// GroupIDFromResourceID extracts the groupID from a hash key made from that groupID
func GroupIDFromResourceID(groupResourceID string) string {
	groupRunes := []rune(groupResourceID)
	return string(groupRunes[2:])
}

// IsStoryReference returns true if the given referenceID (hash key) refers to a story
func IsStoryReference(referenceID string) bool {
	return strings.HasPrefix(referenceID, storyKeyPrefix)
}

// ResourceIDFromP2PEmail creates a hash key for P2P authentication from an email address
func ResourceIDFromP2PEmail(email string) string {
	return p2pKeyPrefix + email
}

// P2PEmailFromResourceID extracts the email from a hash key made from that email
func P2PEmailFromResourceID(p2pResourceID string) string {
	groupRunes := []rune(p2pResourceID)
	return string(groupRunes[2:])
}

// ResourceIDFromStoryID creates a hash key from a storyID
func ResourceIDFromStoryID(storyID string) string {
	return storyKeyPrefix + storyID
}

// ResourceIDFromTransactionID creates a hash key from a transactionID
func ResourceIDFromTransactionID(transactionID string) string {
	return transactionKeyPrefix + transactionID
}

// ReferenceIDFromStoryID creates a hash key from a groupID
func ReferenceIDFromStoryID(storyID string) string {
	return storyKeyPrefix + storyID
}

var sharingTable string

// SetSharingTable allows test scripts to configure a sharing table name without having to
// rely on the environment setting.
func SetSharingTable(sharingTableToSet string) {
	sharingTable = sharingTableToSet
}

// GetTableName returns the name of the table in the given environment
func GetTableName() string {
	if len(sharingTable) > 0 {
		return sharingTable
	}
	return os.Getenv("storyTable")
}

// ----- Keys and Indexes -------

// ResourceIDField holds the hash key of the table
const ResourceIDField = "resourceId"

// ReferenceIDField holds the sort key of the table
const ReferenceIDField = "referenceId"

// UserToGroupIndex is the secondary index that supports querying for all
// groups a user belongs to using their userID
const UserToGroupIndex = "userToGroup"

// EmailIndex is the secondary index that supports querying for the user with
// a specific email address
const EmailIndex = "email"

// PhoneIndex is the secondary index that supports querying for the user with
// a specific phone number
const PhoneIndex = "phone"

// ----- Versioning -------

// VersionField holds the identifier for the latest
// version of a row, used to detect conflicts
const VersionField = "version"

// BaseVersionField holds the identifier for the base
// version of a row, used to detect conflicts
const BaseVersionField = "baseVersion"

// LastUpdatedField holds the milliseconds since epoch when the field was last touched
const LastUpdatedField = "lastUpdated"

// LastUpdatedByField holds the user ID of the last user that updated the
// record
const LastUpdatedByField = "lastUpdatedBy"

// ----- Common fields -------

// IDField holds the primary id for a record without any of the extra formatting
// used to create the ReferenceID or ResourceID
const IDField = "id"

// NameField holds the name of the user, group, etc.
const NameField = "name"

// EmailField holds the email of the user
const EmailField = "email"

// CreatedAtField holds the ms since epoch when the record was created
const CreatedAtField = "createdAt"

// ----- Group & GroupMember fields -------

// GroupIDField holds the identifier for a ShareGroup
const GroupIDField = "groupId"

// OwnerIDField holds the userID for the owner of a group
const OwnerIDField = "ownerId"

// InvitationIDField holds the GroupMember.InvitationID
const InvitationIDField = "invitationId"

// MemberIDField holds the user ID for a group member
const MemberIDField = "memberId"

// MemberEmailField holds the email address of a member
const MemberEmailField = "memberEmail"

// MemberNameField holds the name of a member
const MemberNameField = "memberName"

// InvitedByIDField holds the user ID for the user that invited a member
const InvitedByIDField = "invitedById"

// InvitedByEmailField holds the email address for the user that invited a member
const InvitedByEmailField = "invitedByEmail"

// InvitedByNameField holds the name of the user that invited a member
const InvitedByNameField = "invitedByName"

// InviteAcceptedField holds the current state of the membership acceptance
// can be y|n|p
const InviteAcceptedField = "inviteAccepted"

// InvitedOnField holds the ms since epoch when the invitation was created
const InvitedOnField = "invitedOn"

// ----- Story fields -------

// StoryIDField is the JSON field name for the story ID
const StoryIDField string = "storyID"

// AlbumReferenceField is the JSON field name for the munged title
const AlbumReferenceField string = "albumReference"

// ContentField is the JSON field name for the story content
const ContentField string = "content"

// StorySourceField is the JSON field name for the story source
const StorySourceField string = "storySource"

// StatusField is the JSON field name for the status of a StoryGroup
const StatusField string = "status"

// SharingProductIDField is the JSON field name for the product the user purchased
// to subscribe to sharing
const SharingProductIDField string = "sharingProductId"

// SharingExpiryField is the JSON field name for the date in ms since epoch when the user's
// subscription expires
const SharingExpiryField string = "sharingExpiry"

// ----- Subscription fields -------

// ProductIDField is the JSON field name for the product ID
const ProductIDField string = "productID"

// OriginalTransactionIDField is the JSON field name for the original transaction ID
const OriginalTransactionIDField string = "originalTransactionID"

// TransactionIDField is the JSON field name for the transaction ID
const TransactionIDField string = "transactionID"

// ExpiresMSField is the JSON field name for the expiry time in ms since epoch
const ExpiresMSField string = "expiresMS"

// UserIDField is the JSON field name for the online user ID
const UserIDField string = "userId"

// SharingCompositeKey holds the partition and range keys for the DB.
type SharingCompositeKey struct {
	ResourceID  string `json:"resourceId" dynamodbav:"resourceId"`
	ReferenceID string `json:"referenceId" dynamodbav:"referenceId"`
}
