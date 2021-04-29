package sharing

import (
	"testing"

	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

var testDBData = awsproxy.TestDBData{
	awsproxy.TestDBDataRecord{
		ResourceID:  userReferenceID1,
		ReferenceID: userReferenceID1,
		QueryKey:    email1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:     userReferenceID1,
			ftdb.ReferenceIDField:    userReferenceID1,
			ftdb.EmailField:          email1,
			ftdb.InviteAcceptedField: UserInviteAccepted,
			ftdb.IDField:             userID1,
			ftdb.CreatedAtField:      "1584736061855"},
	},
	awsproxy.TestDBDataRecord{
		ResourceID:  userReferenceID2,
		ReferenceID: userReferenceID2,
		QueryKey:    email2,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:     userReferenceID2,
			ftdb.ReferenceIDField:    userReferenceID2,
			ftdb.EmailField:          email2,
			ftdb.InviteAcceptedField: UserInvitePending,
			ftdb.IDField:             userID2,
			ftdb.CreatedAtField:      "1584736061855"},
	},
	testSubscriptionUser1Record(),
}

func TestNewUserProperties(t *testing.T) {
	newUser := NewOnlineUser(userID1, userName1, email1)
	if newUser.ID != userID1 || newUser.Email != email1 || newUser.Name != userName1 {
		t.Errorf("New user properties not as expected.")
	}
}

func TestNewUserIsAccepted(t *testing.T) {
	newUser := NewOnlineUser(userID1, userName1, email1)
	if !newUser.IsAccepted() || newUser.IsPending() {
		t.Errorf("New user should have been accepted.")
	}
}

func TestInvitedUserProperties(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	if newUser.Email != email1 || newUser.Name != userName1 {
		t.Errorf("Invited user properties not as expected.")
	}
}

func TestInvitedUserIsPending(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	if newUser.IsAccepted() || !newUser.IsPending() {
		t.Errorf("New user should have been pending.")
	}
}

func TestNewUserAllowsEmail(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	if !newUser.AllowsEmail() {
		t.Errorf("New user should allow email.")
	}
}

func TestUserNotAllowsEmailAfterOptOut(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	newUser.EmailOptOut = true
	if newUser.AllowsEmail() {
		t.Errorf("User should not allow email.")
	}
}

func TestHasPhoneWithPhone(t *testing.T) {
	user := OnlineUser{Phone: phone1}
	if !user.HasPhone() {
		t.Errorf("User with phone should have phone.")
	}
}

func TestHasPhoneFalseWithoutPhone(t *testing.T) {
	user := OnlineUser{}
	if user.HasPhone() {
		t.Errorf("User without phone should not have phone.")
	}
}

func TestSubscriptionCheckNoWithNoSubscription(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	if newUser.IsSubscriptionCheckRequired() {
		t.Errorf("User should not need a subscription check.")
	}
}

func TestSubscriptionCheckYesWithSubscriptionAndNoCheck(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: expiry1}
	if !newUser.IsSubscriptionCheckRequired() {
		t.Errorf("User should need a subscription check if they haven't had one.")
	}
}

func TestSubscriptionCheckNoWithSubscriptionAndRecentCheck(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: expiry1, LastExpiryVerify: now()}
	if newUser.IsSubscriptionCheckRequired() {
		t.Errorf("User should not need a subscription check with recent check.")
	}
}

func TestSubscriptionCheckYesWithSubscriptionAndPastDueCheck(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: expiry1, LastExpiryVerify: now() - 2*oneWeek}
	if !newUser.IsSubscriptionCheckRequired() {
		t.Errorf("User should need a subscription check with old verify.")
	}
}

func TestSubscriptionCheckYesWithNearSubscriptionAndPastDueCheck(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: now() + oneHour, LastExpiryVerify: now() - 2*oneHour}
	if !newUser.IsSubscriptionCheckRequired() {
		t.Errorf("User should need a subscription check when near expiry.")
	}
}

func TestSubscriptionExpiredWithNoSubscription(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: 0}
	if !newUser.IsSubscriptionExpired() {
		t.Errorf("Subscription should be expired if they haven't had one.")
	}
}

func TestSubscriptionExpiredFalseWithFutureSubscriptionExpiry(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: now() + oneHour}
	if newUser.IsSubscriptionExpired() {
		t.Errorf("Subscription should not be expired with a future date.")
	}
}

func TestSubscriptionExpiredFalseWithFutureGracePeriod(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: now() - oneHour, GracePeriod: now() + oneHour}
	if newUser.IsSubscriptionExpired() {
		t.Errorf("Subscription should not be expired with a future date.")
	}
}

func TestSubscriptionExpiredWithPastGracePeriod(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: now() - oneHour, GracePeriod: now() - oneHour}
	if !newUser.IsSubscriptionExpired() {
		t.Errorf("Subscription should be expired with a past grace date.")
	}
}

func TestSubscriptionExpiredWithPastSubscriptionExpiry(t *testing.T) {
	newUser := OnlineUser{SharingExpiry: now() - oneHour}
	if !newUser.IsSubscriptionExpired() {
		t.Errorf("Subscription should be expired with a past date.")
	}
}

func TestCanGenerateANewUserID(t *testing.T) {
	newUserID, err := NewTemporaryUserID()
	if nil != err || len(newUserID) == 0 {
		t.Errorf("Did not generate a usable UserID")
	}
}

func TestSetNotificationTokenNoTokensAddsIt(t *testing.T) {
	newUser := OnlineUser{ID: userID1}
	newUser.SetDeviceNotificationToken(appInstallID1, snsEndpoint1, notificationToken1)
	expectFindDeviceNotificationToken(&newUser, appInstallID1, snsEndpoint1, notificationToken1, t)
}

func TestSetNotificationTokenUpdatesExisting(t *testing.T) {
	newUser := testUser1WithNotificationTokens()
	newUser.SetDeviceNotificationToken(appInstallID1, updatedSNSEndpoint1, updatedNotificationToken1)
	expectFindDeviceNotificationToken(&newUser, appInstallID1, updatedSNSEndpoint1, updatedNotificationToken1, t)
}

func TestSetNotificationTokenAddsNewWithExisting(t *testing.T) {
	newUser := testUser1WithNotificationTokens()
	newUser.SetDeviceNotificationToken(appInstallID3, updatedSNSEndpoint1, updatedNotificationToken1)
	expectFindDeviceNotificationToken(&newUser, appInstallID3, updatedSNSEndpoint1, updatedNotificationToken1, t)
}

func TestFindNotificationTokenReturnsNilWithNoTokens(t *testing.T) {
	newUser := OnlineUser{ID: userID1}
	deviceToken := newUser.FindDeviceNotificationToken(appInstallID1)
	if nil != deviceToken {
		t.Error("Expected a nil token")
	}
}

func TestFindNotificationTokenReturnsMatchingToken(t *testing.T) {
	newUser := testUser1WithNotificationTokens()
	expectFindDeviceNotificationToken(&newUser, appInstallID1, snsEndpoint1, notificationToken1, t)
}

func TestSaveWithTokens(t *testing.T) {
	newUser := testUser1WithNotificationTokens()
	expectSaveWithValues(newUser, map[string]interface{}{ftdb.IDField: userID1}, t)
}

func TestLoadOrCreateLoadsForExistingAcceptedUser(t *testing.T) {
	expectCreate(userID1, userName1, email1, true, t)
}

func TestLoadOrCreateLoadsForUserWithSubscription(t *testing.T) {
	user := expectCreate(subscriptionUserID1, subscriptionUserName1, subscriptionUserEmail1, true, t)
	if user.SharingProductID != productID1 || user.SharingExpiry != expiry1 {
		t.Errorf("Product or expiry wrong")
	}
}

func TestLoadOrCreateLoadsForExistingPendingUser(t *testing.T) {
	expectCreate(userID2, userName2, email2, false, t)
}

func TestLoadOrCreateCreatesForNewUser(t *testing.T) {
	expectCreate("", userName3, email3, false, t)
}

func expectCreate(expectedUserID string, userName string, email string, accepted bool, t *testing.T) *OnlineUser {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(expectedUserID, testSvc)
	user, err := LoadOrCreateTemporary(ftCtx, userName, email)
	if nil != err {
		t.Errorf("Expected a user %s", err.Error())
	}
	if nil == user {
		t.Errorf("Should have loaded a user")
	}
	if (user.ID != expectedUserID && len(expectedUserID) > 0) || (accepted && user.IsPending()) ||
		(!accepted && user.IsAccepted()) {
		t.Errorf("Not the expected user")
	}
	return user
}

func TestSaveWorksForNewUser(t *testing.T) {
	newUser := NewOnlineUser(userID1, userName1, email1)
	expectSaveWithValues(newUser, map[string]interface{}{
		ftdb.IDField:             userID1,
		ftdb.EmailField:          email1,
		ftdb.InviteAcceptedField: UserInviteAccepted,
	}, t)
}

func TestSaveWorksForInvitedUser(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	expectSaveWithValues(newUser, map[string]interface{}{
		ftdb.IDField:             userID1,
		ftdb.EmailField:          email1,
		ftdb.InviteAcceptedField: UserInvitePending,
	}, t)
}

func TestSaveFailsForPendingUserWithExistingUser(t *testing.T) {
	newUser := NewOnlineUserInvitation(userID1, userName1, email1)
	expectUserNotSaved(newUser, t)
}

func TestSaveFailsForNewUserWithExistingUser(t *testing.T) {
	newUser := NewOnlineUser(userID1, userName1, email1)
	expectUserNotSaved(newUser, t)
}

func TestDeletesPendingWhenConvertedToAccepted(t *testing.T) {
	newUser := NewOnlineUser(userID2, userName2, email2)
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	err := newUser.Save(awsproxy.NewTestContext(userID2, testSvc))
	if nil != err {
		t.Errorf("Save failed %s", err.Error())
	}
	expectedID := ftdb.ResourceIDFromUserID(userID2)
	testSvc.ExpectDeleteItem(map[string]interface{}{
		ftdb.ResourceIDField:  expectedID,
		ftdb.ReferenceIDField: expectedID,
	}, t)
	testSvc.ExpectPutItem(map[string]interface{}{
		ftdb.IDField:             userID2,
		ftdb.EmailField:          email2,
		ftdb.InviteAcceptedField: UserInviteAccepted,
	}, t)
}

func TestSharingUpdatesForNewSubscription(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	user, err := LoadOrCreateTemporary(ftCtx, userName1, email1)
	err = user.UpdateSharingSubscription(ftCtx, productID1, expiry1, false, expiry1, transactionID1)
	if nil != err {
		t.Errorf("Unexpected error %s", err.Error())
	}
	if user.SharingProductID != productID1 ||
		user.SharingExpiry != expiry1 {
		t.Error("Returned user did not have expected subscription")
	}
}

func TestEmailOptOutUpdatedForUser(t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	ftCtx := awsproxy.NewTestContext(userID1, testSvc)
	user, err := LoadOrCreateTemporary(ftCtx, userName1, email1)
	err = user.Update(ftCtx)
	if nil != err {
		t.Errorf("Unexpected error %s", err.Error())
	}
}

func expectSaveWithValues(newUser OnlineUser, expectedValues map[string]interface{}, t *testing.T) {
	testSvc := awsproxy.NewTestDBSvc()
	err := newUser.Save(awsproxy.NewTestContext(newUser.ID, testSvc))
	if nil != err {
		t.Errorf("Save failed %s", err.Error())
	}
	testSvc.ExpectPutItem(expectedValues, t)
}

func expectUserNotSaved(newUser OnlineUser, t *testing.T) {
	testSvc := awsproxy.NewTestDBSvcWithData(testDBData)
	err := newUser.Save(awsproxy.NewTestContext(newUser.ID, testSvc))
	_, ok := err.(*UserAlreadyExistsError)
	if !ok {
		t.Errorf("Should have failed saving new with an existing user")
	}
	testSvc.ExpectPutCount(0, t)
}

func expectFindDeviceNotificationToken(user *OnlineUser, appInstallID string, snsEndpoint string, notificationToken string, t *testing.T) {
	deviceToken := user.FindDeviceNotificationToken(appInstallID)
	if nil == deviceToken {
		t.Error("Expected a token")
	} else {
		if notificationToken != deviceToken.NotificationToken {
			t.Errorf("Token did not match, expected %s, got %s", notificationToken, deviceToken.NotificationToken)
		}
		if snsEndpoint != deviceToken.SNSEndpoint {
			t.Errorf("Endpoint did not match, expected %s, got %s", snsEndpoint, deviceToken.SNSEndpoint)
		}
	}

}
