package sharing

import (
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftdb"
)

const (
	userID1                      = "user1"
	userID2                      = "user2"
	userID3                      = "user3"
	subscriptionUserID1          = "subscriptionUser1"
	userReferenceID1             = "U#user1"
	userReferenceID2             = "U#user2"
	subscriptionUserReferenceID1 = "U#subscriptionUser1"
	email1                       = "user1@example.com"
	email2                       = "user2@example.com"
	email3                       = "user3@example.com"
	subscriptionUserEmail1       = "subscription1@example.com"
	userName1                    = "username1"
	userName2                    = "username2"
	userName3                    = "username3"
	phone1                       = "+15555555555"
	phone2                       = "+15555555556"
	subscriptionUserName1        = "SubscriptionUserName1"
	appInstallID1                = "app1"
	appInstallID2                = "app2"
	appInstallID3                = "app3"
	snsEndpoint1                 = "sns1"
	snsEndpoint2                 = "sns2"
	updatedSNSEndpoint1          = "updatedSNS1"
	notificationToken1           = "token1"
	notificationToken2           = "token2"
	updatedNotificationToken1    = "updatedToken1"

	// Story Information
	storyID1          = "story1"
	albumReference1   = "albumReference1"
	storyResourceID1  = "S#" + storyID1
	storyReferenceID1 = "S#" + storyID1
	content1          = "Content 1"

	// Group Information
	invitationID1    = "invitation1"
	invitedOn1       = 1583416279780 // milliseconds since epoch
	groupID1         = "group1"
	groupIDMissing   = "groupMissing"
	groupResourceID1 = "G#" + groupID1
	groupName1       = "Group Name 1"
	version1         = "8d3e4d3a-61cd-4877-bd6c-c9d4928565fd"

	// Subscription Information
	productID1     = "product1"
	expiry1        = 1588689735312
	transactionID1 = "1000000659305588"

	newMemberInviteJSON1 = `
	{
		"invitationId":"invitation1",
		"groupId":"group1",
		"memberId":"",
		"memberEmail":"user3@example.com",
		"memberName":"user3@example.com",
		"invitedById":"user1",
		"invitedOn":12038102,
		"inviteAccepted":"P",
		"version":"8d3e4d3a-61cd-4877-bd6c-c9d4928565fd",
		"baseVersion":"",
		"lastUpdated":428409384,
		"lastUpdatedBy":"me",
		"customMsg":"Msg", 
		"groupName":"Group 1"
	}
	`
	shareStoryJSON1 = `
	{
		"groups": [
			{
				"storyId": "story1",
				"groupId": "group1",
				"version": "8d3e4d3a-61cd-4877-bd6c-c9d4928565fe",
				"lastUpdated": 128180131,
				"lastUpdatedBy": "user1",
				"status": "y"
			}
		],
		"sharedStory": {
			"id": "story1",
			"albumReference": "albumReference1",
			"content": "Content 1",
			"version":"8d3e4d3a-61cd-4877-bd6c-c9d4928565fd",
			"lastUpdated":428409384,
			"lastUpdatedBy":"user1",
			"storySource":"gphoto"
		}
	}
	`
	duplicateShareStoryJSON1 = `
	{
		"groups": [
			{
				"storyId": "duplicate1",
				"groupId": "group1",
				"version": "dbf0c989-48a6-4161-a407-c2ecc9e6011d",
				"lastUpdated": 1624909391707,
				"lastUpdatedBy": "user2",
				"status": "y"
			}
		],
		"sharedStory": {
			"id": "duplicate1",
			"albumReference": "albumReference1",
			"content": "Content 2",
			"version":"fa730fcb-815b-44c3-a012-a2f2e6bd2b2b",
			"lastUpdated":1624909391707,
			"lastUpdatedBy":"user2",
			"storySource":"gphoto"
		}
	}
	`
)

func testDeviceNotificationToken1() DeviceNotificationToken {
	return DeviceNotificationToken{AppInstallID: appInstallID1, SNSEndpoint: snsEndpoint1, NotificationToken: notificationToken1}
}

func testDeviceNotificationToken2() DeviceNotificationToken {
	return DeviceNotificationToken{AppInstallID: appInstallID2, SNSEndpoint: snsEndpoint2, NotificationToken: notificationToken2}
}

func testUser1WithNotificationTokens() OnlineUser {
	var deviceTokens = []DeviceNotificationToken{
		testDeviceNotificationToken1(),
		testDeviceNotificationToken2(),
	}
	return OnlineUser{ID: userID1, DeviceTokens: deviceTokens}
}

func testUser1Group1MembershipRecord() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  groupResourceID1,
		ReferenceID: userReferenceID1,
		QueryKey:    userID1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:     groupResourceID1,
			ftdb.ReferenceIDField:    userReferenceID1,
			ftdb.InvitationIDField:   invitationID1,
			ftdb.GroupIDField:        groupID1,
			ftdb.MemberIDField:       userID1,
			ftdb.MemberEmailField:    email1,
			ftdb.MemberNameField:     userName1,
			ftdb.InvitedByIDField:    userID2,
			ftdb.InvitedByNameField:  userName2,
			ftdb.InvitedByEmailField: email2,
			ftdb.InvitedOnField:      1584817936287,
			ftdb.InviteAcceptedField: MembershipAccepted,
			ftdb.VersionField:        version1,
			ftdb.BaseVersionField:    version1,
			ftdb.LastUpdatedField:    1584817936287,
			ftdb.LastUpdatedByField:  userName1,
		},
	}

}

func testUser1Record() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  userReferenceID1,
		ReferenceID: userReferenceID1,
		QueryKey:    email1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:     userReferenceID1,
			ftdb.ReferenceIDField:    userReferenceID1,
			ftdb.IDField:             userID1,
			ftdb.EmailField:          email1,
			ftdb.InviteAcceptedField: UserInviteAccepted,
			ftdb.CreatedAtField:      "1280913280",
		},
	}
}

func testGroup1Record() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  groupResourceID1,
		ReferenceID: groupResourceID1,
		QueryKey:    groupID1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:  groupResourceID1,
			ftdb.ReferenceIDField: groupResourceID1,
			ftdb.IDField:          groupID1,
			ftdb.NameField:        groupName1,
		},
	}
}

func testSubscriptionUser1Record() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  subscriptionUserReferenceID1,
		ReferenceID: subscriptionUserReferenceID1,
		QueryKey:    subscriptionUserEmail1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:       subscriptionUserReferenceID1,
			ftdb.ReferenceIDField:      subscriptionUserReferenceID1,
			ftdb.IDField:               subscriptionUserID1,
			ftdb.EmailField:            subscriptionUserEmail1,
			ftdb.InviteAcceptedField:   UserInviteAccepted,
			ftdb.CreatedAtField:        "1280913280",
			ftdb.SharingProductIDField: productID1,
			ftdb.SharingExpiryField:    expiry1,
		},
	}
}

func testStory1Record() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  storyResourceID1,
		ReferenceID: storyReferenceID1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:     storyResourceID1,
			ftdb.ReferenceIDField:    storyReferenceID1,
			ftdb.IDField:             storyID1,
			ftdb.AlbumReferenceField: albumReference1,
			ftdb.ContentField:        content1,
			ftdb.VersionField:        version1,
			ftdb.BaseVersionField:    version1,
			ftdb.LastUpdatedField:    1584817936287,
			ftdb.LastUpdatedByField:  userName1,
		},
	}
}

func testStory1Group1Record() awsproxy.TestDBDataRecord {
	return awsproxy.TestDBDataRecord{
		ResourceID:  groupResourceID1,
		ReferenceID: storyReferenceID1,
		QueryKey:    groupResourceID1,
		Record: map[string]interface{}{
			ftdb.ResourceIDField:    groupResourceID1,
			ftdb.ReferenceIDField:   storyReferenceID1,
			ftdb.GroupIDField:       groupID1,
			ftdb.StoryIDField:       storyID1,
			ftdb.VersionField:       version1,
			ftdb.BaseVersionField:   version1,
			ftdb.LastUpdatedField:   1584817936287,
			ftdb.LastUpdatedByField: userName1,
			ftdb.StatusField:        StoryGroupActive,
		},
	}
}
