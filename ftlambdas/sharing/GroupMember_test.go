package sharing

import (
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
	"github.com/stretchr/testify/assert"
)

var groupTestDBData = awsproxy.TestDBData{
	testUser1Group1MembershipRecord(),
	testGroup1Record(),
	testUser1Record(),
}

func TestEmptyGroupMemberIsNotNewInvite(t *testing.T) {
	member := GroupMember{}
	if member.IsNewMemberInvite() {
		t.Errorf("Empty member should not qualify")
	}
}

func TestEmptyGroupMemberIsInvalid(t *testing.T) {
	member := GroupMember{}
	if !member.IsInvalid() {
		t.Errorf("Empty member should be invalid")
	}
}

func TestEmptyMemberIDIsNewInvite(t *testing.T) {
	member := GroupMember{
		InvitationID: invitationID1,
		MemberEmail:  email1,
		InvitedByID:  userID1,
		InvitedOn:    invitedOn1,
		GroupID:      groupID1,
	}
	if !member.IsNewMemberInvite() {
		t.Errorf("Member should qualify")
	}
}

func TestNonEmptyMemberIDIsNotNewInvite(t *testing.T) {
	member := GroupMember{
		MemberID:     userID1,
		InvitationID: invitationID1,
		MemberEmail:  email1,
		InvitedByID:  userID1,
		InvitedOn:    invitedOn1,
		GroupID:      groupID1,
	}
	if member.IsNewMemberInvite() {
		t.Errorf("Has memberID, should not be new")
	}
}

func TestEmptyIsNotValidUpdate(t *testing.T) {
	member := GroupMember{}
	if member.IsValidUpdate() {
		t.Errorf("Empty member should not qualify")
	}
}

func TestPopulatedMemberIsValidUpdate(t *testing.T) {
	member := GroupMember{
		MemberID:       userID1,
		InvitationID:   invitationID1,
		MemberEmail:    email1,
		InvitedByID:    userID1,
		InvitedOn:      invitedOn1,
		GroupID:        groupID1,
		InviteAccepted: MembershipAccepted,
		Version:        version1,
	}
	if !member.IsValidUpdate() {
		t.Errorf("Member should qualify")
	}
}

func TestWithoutVersionIsNotValidUpdate(t *testing.T) {
	member := GroupMember{
		MemberID:       userID1,
		InvitationID:   invitationID1,
		MemberEmail:    email1,
		InvitedByID:    userID1,
		InvitedOn:      invitedOn1,
		GroupID:        groupID1,
		InviteAccepted: MembershipAccepted,
	}
	if member.IsValidUpdate() {
		t.Errorf("Member should require version")
	}
}

func TestIncorrectAcceptedIsNotValidUpdate(t *testing.T) {
	member := GroupMember{
		MemberID:       userID1,
		InvitationID:   invitationID1,
		MemberEmail:    email1,
		InvitedByID:    userID1,
		InvitedOn:      invitedOn1,
		GroupID:        groupID1,
		InviteAccepted: "junk",
	}
	if member.IsValidUpdate() {
		t.Errorf("Member should validate InviteAccepted in y|n|r")
	}
}

func TestLoadMemberWorksForExistingMember(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	member, err := LoadGroupMember(ftCtx, groupID1, userID1)
	assert.NoError(t, err)
	if member.InvitationID != invitationID1 ||
		member.MemberID != userID1 ||
		member.InviteAccepted != MembershipAccepted {
		t.Errorf("Loaded member does not have expected values")
	}
}

func TestSwitchUserChangesTheCorrectMember(t *testing.T) {
	testDB, ftCtx := createTestContext(userID2)
	err := SwitchAllMembersToPermanentUser(ftCtx, userID2, userID1)
	assert.NoError(t, err)
	testDB.ExpectDeleteItem(map[string]interface{}{
		ftdb.ResourceIDField:  ftdb.ResourceIDFromGroupID(groupID1),
		ftdb.ReferenceIDField: ftdb.ReferenceIDFromUserID(userID1),
	}, t)
	testDB.ExpectPutItem(map[string]interface{}{
		ftdb.InvitationIDField: invitationID1,
		ftdb.MemberIDField:     userID2,
	}, t)
}

func TestUpdateSendsMailForNewInvite(t *testing.T) {
	_, ftCtx := createTestContext(userID2)
	err := UpdateGroupMembership(ftCtx, newMemberInviteJSON1)
	assert.NoError(t, err)
	expectContent := awsproxy.EmailContent{
		Subject: email1 + " invites you to join Folktells",
	}
	emailSvc := ftCtx.EmailSvc
	testSvc, ok := emailSvc.(*awsproxy.TestEmailSender)
	if ok {
		testSvc.ExpectLastEmailContent(expectContent, t)
	}
}

func TestFindGroupsForUserFindsCorrectGroup(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	groups, err := FindGroupsForUser(ftCtx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, groupID1, groups[0])
}

func TestFindGroupsForUserHandlesNoGroups(t *testing.T) {
	_, ftCtx := createTestContext(userID2)
	groups, err := FindGroupsForUser(ftCtx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(groups))
}
func TestFindMembersForGroupHandlesNoGroup(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	groups, err := FindMembersForGroup(ftCtx, groupIDMissing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(groups.Members))
}

func TestFindMembersForGroupFindsExpectedMembers(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	groups, err := FindMembersForGroup(ftCtx, groupID1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(groups.Members))
}

func TestFindOnlineUsersForGroupHandlesEmptyGroup(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	users, err := FindOnlineUsersForGroup(ftCtx, groupIDMissing)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(users))
}

func TestFindOnlineUsersForGroupFindsExpectedUser(t *testing.T) {
	_, ftCtx := createTestContext(userID1)
	users, err := FindOnlineUsersForGroup(ftCtx, groupID1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))
}

func createTestContext(userID string) (*awsproxy.TestDynamoDB, awsproxy.FTContext) {
	testDB := awsproxy.NewTestDBSvcWithData(groupTestDBData)
	ftCtx := awsproxy.NewTestContext(userID, testDB)
	return testDB, ftCtx
}
