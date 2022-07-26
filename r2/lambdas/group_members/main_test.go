package main

import (
	"encoding/json"
	"testing"

	"github.com/sowens-csd/folktells-server/sharing"
)

func TestRequestForAllGroups(t *testing.T) {
	svc := &stubDynamoDB{}
	params := make(map[string]string)
	response := findMatchingGroups(params, "request1", "user1", svc)
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status")
	} else {
		groupMembersResp := parseMembersJSON(response.Body, t)
		if len(groupMembersResp.Groups) != 1 {
			t.Errorf("No group")
		}
	}
}

func TestRequestForParticularGroup(t *testing.T) {
	svc := &stubDynamoDB{}
	params := make(map[string]string)
	params["groupID"] = "GroupId1"
	response := findMatchingGroups(params, "request1", "user1", svc)
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status")
	} else {
		groupMembersResp := parseMembersJSON(response.Body, t)
		if len(groupMembersResp.Groups) != 1 {
			t.Errorf("No group")
		}
	}
}

func TestFindSingleGroup(t *testing.T) {
	svc := &stubDynamoDB{}
	group, err := createGroup("GroupId1", svc, log.WithFields(log.Fields{}))
	if nil != err {
		t.Errorf("Failed creating group %s", err.Error())
	}
	if group.GroupID != "Group1" {
		t.Errorf("Expected group not found")
	}
}

func TestFindGroupMembersForSingleGroup(t *testing.T) {
	svc := &stubDynamoDB{}
	members, err := membersForGroup("GroupId1", svc, log.WithFields(log.Fields{}))
	if nil != err {
		t.Errorf("Failed creating group %s", err.Error())
	}
	if members.Group.GroupID != "Group1" {
		t.Errorf("Expected group not found")
	}
	if len(members.Members) != 1 {
		t.Errorf("Expected members not found")
	}
}

func TestBuildGroupMembersResponseForSingleGroup(t *testing.T) {
	var members []sharing.GroupMember
	members = append(members, sharing.GroupMember{
		InvitationID: "invitation1",
		MemberID:     "user2",
	})
	var groupMembers []GroupMembers
	groupMembers = append(groupMembers, GroupMembers{
		Group: sharing.ShareGroup{
			GroupID: "Group1",
			Name:    "Group 1",
			OwnerID: "user1",
		},
		Members: members,
	})
	response := groupMembersResponse(groupMembers, log.WithFields(log.Fields{}))
	if response.StatusCode != 200 {
		t.Errorf("Unexpected status")
	}
	groupMembersResp := parseMembersJSON(response.Body, t)
	if len(groupMembersResp.Groups) != 1 {
		t.Errorf("No group")
	}
}

func parseMembersJSON(body string, t *testing.T) GroupMembersResponse {
	var members GroupMembersResponse
	membersJSON := []byte(body)
	err := json.Unmarshal(membersJSON, &members)
	if err != nil {
		t.Errorf(err.Error())
	}
	return members
}
