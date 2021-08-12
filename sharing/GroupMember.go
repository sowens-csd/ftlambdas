package sharing

import (
	"bytes"
	"encoding/json"
	"fmt"
	htmltmpl "html/template"
	"strings"
	texttmpl "text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	uuid "github.com/satori/go.uuid"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

// GroupMember is the information exchanged for a single membership
// record between the server and clients
type GroupMember struct {
	InvitationID   string `json:"invitationId" dynamodbav:"invitationId,omitempty"`
	GroupID        string `json:"groupId" dynamodbav:"groupId"`
	MemberID       string `json:"memberId" dynamodbav:"memberId"`
	MemberEmail    string `json:"memberEmail" dynamodbav:"memberEmail"`
	MemberName     string `json:"memberName,omitempty" dynamodbav:"memberName"`
	InvitedByID    string `json:"invitedById" dynamodbav:"invitedById"`
	InvitedByName  string `json:"invitedByName,omitempty" dynamodbav:"invitedByName"`
	InvitedByEmail string `json:"invitedByEmail" dynamodbav:"invitedByEmail"`
	InvitedOn      int    `json:"invitedOn" dynamodbav:"invitedOn"`
	InviteAccepted string `json:"inviteAccepted" dynamodbav:"inviteAccepted"`
	Version        string `json:"version" dynamodbav:"version"`
	BaseVersion    string `json:"baseVersion" dynamodbav:"baseVersion"`
	LastUpdated    int    `json:"lastUpdated" dynamodbav:"lastUpdated"`
	LastUpdatedBy  string `json:"lastUpdatedBy,omitempty" dynamodbav:"lastUpdatedBy,omitempty"`
	GroupName      string `json:"groupName,omitempty" dynamodbav:"groupName,omitempty"`
	CustomMsg      string `json:"customMsg,omitempty" dynamodbav:"customMsg,omitempty"`
}

// GroupMembers is the content for a set of members
type GroupMembers struct {
	Group   ShareGroup    `json:"group" dynamodbav:"group"`
	Members []GroupMember `json:"members" dynamodbav:"members"`
}

type userToGroupIndexQueryResult struct {
	ResourceID string `json:"resourceId" dynamodbav:"resourceId"`
	MemberID   string `json:"memberId" dynamodbav:"memberId"`
}

type inviteEmailParameters struct {
	GroupName      string
	InvitedByEmail string
	CustomMsg      string
}

// MembershipAccepted is the value for InviteAccepted when the member has
// agreed to be part of the group.
const MembershipAccepted = "y"

// MembershipPending is the value for InviteAccepted when the inviation has
// been extended but not yet accepted or declined.
const MembershipPending = "p"

// MembershipDeclined is the value for InviteAccepted when the member has
// declined to be part of the group.
const MembershipDeclined = "n"

// MembershipRemoved is the value for InviteAccepted when the member has
// been removed from the group.
const MembershipRemoved = "r"

const notFoundGroupID = "notFoundGroupId"

var memberConfirmations map[string]map[string]bool

// UpdateGroupMembership ensures that the member record is properly updated given
// the JSON that is provided.
func UpdateGroupMembership(ctx awsproxy.FTContext, membershipJSON string) error {
	inputJSON := []byte(membershipJSON)
	var groupMember GroupMember
	err := json.Unmarshal(inputJSON, &groupMember)
	if nil != err {
		return err
	}
	newMemberInvite, newUser, emailOptOut, err := groupMember.Save(ctx)
	if nil == err && newMemberInvite && !emailOptOut {
		params := inviteEmailParameters{
			CustomMsg:      groupMember.CustomMsg,
			GroupName:      groupMember.GroupName,
			InvitedByEmail: groupMember.InvitedByEmail,
		}
		if len(groupMember.GroupName) == 0 || len(groupMember.InvitedByEmail) == 0 {
			ctx.RequestLogger.Info().Msg("Group name or email unexpectedly empty.")
		}
		emailContent := englishNewMemberInvite()
		if newUser {
			emailContent = englishNewUserInvite()
		}
		populatedContent, mailErr := generateEmailFromTemplate(params, emailContent)
		if nil == mailErr && nil != populatedContent {
			ctx.EmailSvc.SendEmail(ctx.Context, groupMember.MemberEmail, *populatedContent, ctx.RequestLogger)
		} else if nil != mailErr {
			return mailErr
		}
	}
	return err
}

// UpdateMemberName update all group membership records to reflect the new
// member name.
func UpdateMemberName(ftCtx awsproxy.FTContext, name string) error {
	type memberUpdate struct {
		Name string `json:":n" dynamodbav:":n"`
	}
	groups, err := FindGroupsForUser(ftCtx)
	if nil != err {
		return err
	}
	referenceID := ftdb.ReferenceIDFromUserID(ftCtx.UserID)
	for _, groupID := range groups {
		update, err := attributevalue.MarshalMap(memberUpdate{
			Name: name,
		})
		if err != nil {
			return err
		}
		resourceID := ftdb.ResourceIDFromGroupID(groupID)
		_, err = ftCtx.DBSvc.UpdateItem(ftCtx.Context, &dynamodb.UpdateItemInput{
			TableName: aws.String(ftdb.GetTableName()),
			Key: map[string]types.AttributeValue{
				ftdb.ResourceIDField:  &types.AttributeValueMemberS{Value: resourceID},
				ftdb.ReferenceIDField: &types.AttributeValueMemberS{Value: referenceID},
			},
			UpdateExpression:          aws.String("set memberName = :n"),
			ExpressionAttributeValues: update,
		})
	}
	return nil
}

// AreInSameGroup returns true if the current user and the supplied user are both
// active members of the same group
func AreInSameGroup(ftCtx awsproxy.FTContext, otherUserID string) (bool, error) {
	if nil == memberConfirmations {
		ftCtx.RequestLogger.Debug().Msg("AreInSameGroup creating confirmations map.")
		memberConfirmations = make(map[string]map[string]bool)
	}
	if nil == memberConfirmations[ftCtx.UserID] {
		ftCtx.RequestLogger.Debug().Str("UserID", ftCtx.UserID).Msg("AreInSameGroup adding to confirmation map.")
		memberConfirmations[ftCtx.UserID] = make(map[string]bool)
	}
	ftCtx.RequestLogger.Debug().Str("otherUserID", otherUserID).Msg("AreInSameGroup looking up in confirmation map.")
	inSame, ok := memberConfirmations[ftCtx.UserID][otherUserID]
	if ok {
		ftCtx.RequestLogger.Debug().Bool("otherUserID", inSame).Msg("confirmation map hit unlocking")
		return inSame, nil
	}

	ftCtx.RequestLogger.Debug().Str("currentUser", ftCtx.UserID).Str("otherUser", otherUserID).Msg("confirmation map miss")
	inSame = false
	currentUserGroups, err := findMembersForUser(ftCtx, ftCtx.UserID)
	if err != nil {
		return false, err
	}
	otherUserGroups, err := findMembersForUser(ftCtx, otherUserID)
	if err != nil {
		return false, err
	}
	ftCtx.RequestLogger.Debug().Msg("AreInSameGroup looping groups.")
	for _, currentMember := range currentUserGroups {
		ftCtx.RequestLogger.Debug().Str("group", currentMember.ResourceID).Str("member", currentMember.MemberID).Msg("loading current membership")
		currentGroupMember, err := LoadGroupMember(ftCtx, ftdb.GroupIDFromResourceID(currentMember.ResourceID), currentMember.MemberID)
		if nil == err && currentGroupMember.IsActive() {
			ftCtx.RequestLogger.Debug().Str("group", currentGroupMember.GroupID).Msg("active checking group")
			for _, otherMember := range otherUserGroups {
				ftCtx.RequestLogger.Debug().Str("group", otherMember.ResourceID).Str("member", otherMember.MemberID).Msg("loading other membership")
				otherGroupMember, err := LoadGroupMember(ftCtx, ftdb.GroupIDFromResourceID(otherMember.ResourceID), otherMember.MemberID)
				if nil != err {
					ftCtx.RequestLogger.Info().Err(err).Msg("error loading membership")
				} else {
					ftCtx.RequestLogger.Debug().Str("group", otherGroupMember.GroupID).Str("member", otherGroupMember.MemberID).Str("accepted", otherGroupMember.InviteAccepted).Msg("loaded other membership")
				}
				if nil == err && otherGroupMember.IsActive() && otherMember.ResourceID == currentMember.ResourceID {
					ftCtx.RequestLogger.Debug().Msg("found active membership.")
					inSame = true
					break
				}
			}
		}
		if inSame {
			break
		}
	}

	ftCtx.RequestLogger.Debug().Bool("group", inSame).Msg("updating map.")
	memberConfirmations[ftCtx.UserID][otherUserID] = inSame
	return inSame, nil
}

// DeleteGroupsForUser deletes all the group this user owns and removes them as members from all groups
func DeleteGroupsForUser(ftCtx awsproxy.FTContext) error {
	groups, err := FindGroupsForUser(ftCtx)
	if nil == err {
		userID := ftCtx.UserID
		for _, group := range groups {
			shareGroup, err := LoadGroup(ftCtx, group)
			if nil == err {
				noOtherMembers := true
				memberships, err := findMembersForGroup(ftCtx, group)
				if nil != err {
					ftCtx.RequestLogger.Info().Str("groupID", group).Err(err).Msg("could not find members for group")
				} else {
					for _, member := range memberships {
						if member.MemberID != userID && (member.InviteAccepted == MembershipAccepted || member.InviteAccepted == MembershipPending) {
							noOtherMembers = false
						}
						if member.MemberID == userID && member.InviteAccepted != MembershipRemoved {
							groupMember, err := LoadGroupMember(ftCtx, group, userID)
							if nil != err {
								ftCtx.RequestLogger.Info().Str("groupID", group).Err(err).Msg("could not load membership")
							} else {
								groupMember.InviteAccepted = MembershipRemoved
								newVersion := uuid.NewV4().String()
								groupMember.Version = newVersion
								groupMember.putMember(ftCtx)
							}
						}
					}
				}
				if noOtherMembers {
					err = shareGroup.Delete(ftCtx)
					if nil != err {
						ftCtx.RequestLogger.Info().Str("groupID", group).Err(err).Msg("could not delete group")
					}
				}
			} else {
				ftCtx.RequestLogger.Info().Str("groupID", group).Err(err).Msg("could not load group")
			}
		}
	} else {
		ftCtx.RequestLogger.Info().Err(err).Msg("could not find groups for user")
		return err
	}
	return nil
}

// FindGroupsForUser returns an array of GroupIDs that the given user is an active member of
func FindGroupsForUser(ctx awsproxy.FTContext) ([]string, error) {
	// TODO this is currently looking up all members, not active members
	userGroupResults, err := findMembersForUser(ctx, ctx.UserID)
	if err != nil {
		return nil, err
	}

	var userGroups []string
	for _, userToGroup := range userGroupResults {
		ctx.RequestLogger.Debug().Str("group", ftdb.GroupIDFromResourceID(userToGroup.ResourceID)).Msg("group for user")
		userGroups = append(userGroups, ftdb.GroupIDFromResourceID(userToGroup.ResourceID))
	}
	return userGroups, nil
}

// FindOnlineUsersForGroup load the online users corresponding to the members of a group
func FindOnlineUsersForGroup(ftCtx awsproxy.FTContext, groupID string) ([]*OnlineUser, error) {
	var users []*OnlineUser
	members, err := FindMembersForGroup(ftCtx, groupID)
	if nil != err {
		return users, err
	}

	for _, member := range members.Members {
		ou, err := LoadOnlineUser(ftCtx, member.MemberID)
		if nil == err {
			users = append(users, ou)
		} else {
			ftCtx.RequestLogger.Info().Err(err).Str("uid", member.MemberID).Msg("Failed to load")
		}
	}
	return users, nil
}

// FindMembersForGroup find all the group members that are members of the given group
func FindMembersForGroup(ftCtx awsproxy.FTContext, groupID string) (*GroupMembers, error) {
	group, err := createGroup(ftCtx, groupID)
	if nil != err {
		return nil, err
	}
	results, err := findMembersForGroup(ftCtx, groupID)
	if nil != err {
		return nil, err
	}
	members, err := createGroupMembers(ftCtx, results)
	if nil != err {
		return nil, err
	}

	return &GroupMembers{
		Group:   *group,
		Members: members,
	}, nil
}

func createGroup(ftCtx awsproxy.FTContext, groupID string) (*ShareGroup, error) {
	resourceID := ftdb.ResourceIDFromGroupID(groupID)
	var shareGroup ShareGroup
	_, err := ftdb.GetItem(ftCtx, resourceID, resourceID, &shareGroup)
	if nil != err {
		return nil, err
	}
	return &shareGroup, nil
}

func findMembersForGroup(ftCtx awsproxy.FTContext, groupID string) ([]ftdb.FolktellsRecord, error) {
	ftCtx.RequestLogger.Info().Msg("Querying")
	results, err := ftdb.QueryByResource(ftCtx, ftdb.ResourceIDFromGroupID(groupID))
	ftCtx.RequestLogger.Info().Int("rows", len(results)).Msg("DB result")
	return results, err
}

func createGroupMembers(ftCtx awsproxy.FTContext, queryResults []ftdb.FolktellsRecord) ([]GroupMember, error) {

	var members []GroupMember
	ftCtx.RequestLogger.Info().Int("rows", len(queryResults)).Msg("Creating response")
	for _, result := range queryResults {
		if ftdb.IsUserReference(result.ReferenceID) {
			ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Member of group %s", result.GroupID))
			lastUpd := result.LastUpdated
			now := time.Now()
			sec := now.UTC().Unix()
			lastUpdated := int(sec * 1000)
			if lastUpd > 0 {
				lastUpdated = lastUpd
			}
			members = append(members, GroupMember{
				InvitationID:   result.InvitationID,
				GroupID:        result.GroupID,
				MemberID:       result.MemberID,
				MemberEmail:    result.MemberEmail,
				MemberName:     result.MemberName,
				InvitedByID:    result.InvitedByID,
				InvitedByEmail: result.InvitedByEmail,
				InviteAccepted: result.InviteAccepted,
				Version:        result.Version,
				BaseVersion:    result.BaseVersion,
				LastUpdated:    lastUpdated,
				LastUpdatedBy:  result.LastUpdatedBy,
			})
		}
	}

	return members, nil
}

// Save updates the db with the current groupMember state
// Returns isNewMemberInvite, isNewUser, isEmailOptOut, error
func (groupMember *GroupMember) Save(ctx awsproxy.FTContext) (bool, bool, bool, error) {
	if groupMember.IsInvalid() {
		ctx.RequestLogger.Debug().Str("memberID", groupMember.MemberID).Str(
			"invitationID", groupMember.InvitationID).Str(
			"MemberEmail", groupMember.MemberEmail).Str(
			"MemberName", groupMember.MemberName).Str(
			"InvitedByID", groupMember.InvitedByID).Str(
			"groupID", groupMember.GroupID).Str(
			"inviteAccepted", groupMember.InviteAccepted).Str(
			"version", groupMember.Version).Msg("Invalid groupMember, leaving")
		return false, false, false, fmt.Errorf("Cannot update invitation")
	}
	newMemberInvite := groupMember.IsNewMemberInvite()
	newUser := false
	emailOptOut := false
	if newMemberInvite {
		ctx.RequestLogger.Debug().Msg("New member invite")
		ou, err := LoadOrCreateTemporary(ctx, groupMember.MemberEmail, groupMember.MemberEmail)
		if err != nil {
			return false, false, false, err
		}
		ctx.RequestLogger.Debug().Msg("Loading user")
		groupMember.MemberID = ou.ID
		iu, err := LoadOnlineUser(ctx, groupMember.InvitedByID)
		if err != nil {
			return false, false, false, err
		}
		ctx.RequestLogger.Debug().Msg("Got invite")
		groupMember.InvitedByEmail = iu.Email
		newUser = ou.IsPending()
		emailOptOut = ou.EmailOptOut
	}
	err := groupMember.putMember(ctx)
	if nil != err {
		return false, false, false, err
	}
	return newMemberInvite, newUser, emailOptOut, nil
}

// IsActive returns true if the member is accepted and not deleted.
func (groupMember *GroupMember) IsActive() bool {
	return groupMember.InviteAccepted == MembershipAccepted
}

// IsInvalid returns true if the GroupMember record should not be
// processed because it is not valid.
func (groupMember *GroupMember) IsInvalid() bool {
	return !groupMember.IsNewMemberInvite() && !groupMember.IsValidUpdate()
}

// IsNewMemberInvite returns true if the GroupMember record has the expected
// information for a new member invite.
func (groupMember *GroupMember) IsNewMemberInvite() bool {
	return len(groupMember.MemberID) == 0 &&
		len(groupMember.InvitationID) > 0 &&
		len(groupMember.MemberEmail) > 0 &&
		len(groupMember.InvitedByID) > 0 &&
		groupMember.InvitedOn > 0 &&
		len(groupMember.GroupID) > 0
}

// IsValidUpdate returns true if the GroupMember record has the expected
// information for an update to a membership record.
func (groupMember *GroupMember) IsValidUpdate() bool {
	valid := len(groupMember.MemberID) > 0 &&
		len(groupMember.InvitationID) > 0 &&
		len(groupMember.MemberEmail) > 0 &&
		len(groupMember.InvitedByID) > 0 &&
		len(groupMember.GroupID) > 0 &&
		len(groupMember.InviteAccepted) > 0 &&
		len(groupMember.Version) > 0
	switch groupMember.InviteAccepted {
	case
		MembershipAccepted,
		MembershipDeclined,
		MembershipPending,
		MembershipRemoved:
		break
	default:
		valid = false
	}
	return valid
}

// ResourceID returns the ResourceID for this member
func (groupMember *GroupMember) ResourceID() string {
	return ftdb.ReferenceIDFromGroupID(groupMember.GroupID)
}

// ReferenceID returns the ReferenceID for this member
func (groupMember *GroupMember) ReferenceID() string {
	return ftdb.ReferenceIDFromUserID(groupMember.MemberID)
}

// LoadGroupMember attempt to load a group member given the group and user, always returns a GroupMember
// check the .IsNotFound property.
func LoadGroupMember(ctx awsproxy.FTContext, groupID string, userID string) (*GroupMember, error) {
	resourceID := ftdb.ResourceIDFromGroupID(groupID)
	referenceID := ftdb.ReferenceIDFromUserID(userID)
	groupMember := GroupMember{}
	found, err := ftdb.GetItem(ctx, resourceID, referenceID, &groupMember)
	if err != nil {
		return nil, err
	}
	if !found {
		return &GroupMember{
			GroupID: notFoundGroupID,
		}, nil
	}
	return &groupMember, nil
}

// SwitchAllMembersToPermanentUser moves all group records that used to refer to a temporary user ID and
// updates them to refer to the new permanent user ID.
func SwitchAllMembersToPermanentUser(ctx awsproxy.FTContext, newUserID string, temporaryID string) error {
	membersToUpdate, err := findMembersForUser(ctx, temporaryID)
	if err != nil {
		return err
	}
	for _, memberToUpdate := range membersToUpdate {
		groupID := ftdb.GroupIDFromResourceID(memberToUpdate.ResourceID)
		groupMember, err := LoadGroupMember(ctx, groupID, memberToUpdate.MemberID)
		if err != nil {
			ctx.RequestLogger.Info().Str(
				"resourceId", memberToUpdate.ResourceID).Str(
				"memberID", memberToUpdate.MemberID).Msg("Load member failed.")
		}
		groupMember.SwitchUserID(ctx, newUserID)
	}
	fmt.Println(membersToUpdate)
	// Delete each temporary
	// Add back its permanent counterpart
	return nil
}

func (groupMember *GroupMember) putMember(ctx awsproxy.FTContext) error {
	ctx.RequestLogger.Debug().Msg("Putting invite")
	condtionExp := " attribute_not_exists(resourceId) or inviteAccepted=:inviteRemoved"
	var expressionValues map[string]types.AttributeValue = make(map[string]types.AttributeValue)
	expressionValues[":inviteRemoved"] = &types.AttributeValueMemberS{Value: "r"}
	if len(groupMember.BaseVersion) > 0 {
		condtionExp = condtionExp + " or baseVersion=:baseVersion"
		expressionValues[":baseVersion"] = &types.AttributeValueMemberS{Value: groupMember.BaseVersion}
	}
	groupMember.BaseVersion = groupMember.Version
	resourceID := ftdb.ResourceIDFromGroupID(groupMember.GroupID)
	referenceID := ftdb.ReferenceIDFromUserID(groupMember.MemberID)
	err := ftdb.PutItem(ctx, resourceID, referenceID, groupMember)
	if nil != err {
		ctx.RequestLogger.Error().Msg(fmt.Sprintf("save group member failed %s", err.Error()))
	} else {
		ctx.RequestLogger.Info().Msg("save group member succeeded")
	}
	return err
}

// Delete removes the groupMember from the db
func (groupMember *GroupMember) Delete(ctx awsproxy.FTContext) error {
	return ftdb.DeleteItem(ctx, groupMember.ResourceID(), groupMember.ReferenceID())
}

// SwitchUserID changes the group member record to a different user ID. This is
// used when a pending user switches to a permanent ID.
func (groupMember *GroupMember) SwitchUserID(ctx awsproxy.FTContext, newUserID string) error {
	err := groupMember.Delete(ctx)
	if nil != err {
		return err
	}
	groupMember.MemberID = newUserID
	_, _, _, err = groupMember.Save(ctx)
	return err
}

func findMembersForUser(ctx awsproxy.FTContext, userID string) ([]userToGroupIndexQueryResult, error) {
	result, err := ctx.DBSvc.Query(ctx.Context, &dynamodb.QueryInput{
		TableName:              aws.String(ftdb.GetTableName()),
		IndexName:              aws.String(ftdb.UserToGroupIndex),
		KeyConditionExpression: aws.String("memberId = :userID "),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userID": &types.AttributeValueMemberS{Value: userID},
		},
	})
	userResults := []userToGroupIndexQueryResult{}
	if err != nil {
		return userResults, err
	} else if result.Count == 0 {
		return userResults, nil
	}
	err = attributevalue.UnmarshalListOfMaps(result.Items, &userResults)
	if nil != err {
		return nil, err
	}
	return userResults, nil
}

func generateEmailFromTemplate(params inviteEmailParameters, templates awsproxy.EmailContent) (*awsproxy.EmailContent, error) {
	subject, err := processEmailTemplate(params, templates.Subject)
	if nil != err {
		return nil, err
	}
	htmlBody, htmlErr := processHTMLEmailTemplate(params, templates.HTMLBody)
	if nil != htmlErr {
		return nil, htmlErr
	}
	textBody, txtErr := processEmailTemplate(params, templates.TextBody)
	if nil != txtErr {
		return nil, txtErr
	}
	return &awsproxy.EmailContent{
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	}, nil
}

func processEmailTemplate(params inviteEmailParameters, template string) (string, error) {
	tmpl, err := texttmpl.New("test").Parse(template)
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	err = tmpl.Execute(&result, params)
	return result.String(), err
}

func processHTMLEmailTemplate(params inviteEmailParameters, template string) (string, error) {
	tmpl, err := htmltmpl.New("test").Parse(template)
	if err != nil {
		return "", err
	}
	var result bytes.Buffer
	err = tmpl.Execute(&result, params)
	return result.String(), err
}

func isInvitation(membershipJSON string) bool {
	return strings.Contains(membershipJSON, "\"customMsg\":")
}
