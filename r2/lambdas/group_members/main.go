package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/sharing"
)

// GroupMembersResponse supports paging data across multiple calls
type GroupMembersResponse struct {
	PageToken string
	Groups    []sharing.GroupMembers `json:"groups"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	return findMatchingGroups(ftCtx, request.PathParameters), nil
}

func findMatchingGroups(ftCtx awsproxy.FTContext, params map[string]string) awsproxy.Response {
	rawGroupID, found := params["groupID"]
	if found && len(rawGroupID) > 0 {
		groupID, err := url.QueryUnescape(rawGroupID)
		if nil != err {
			return awsproxy.HandleError(err, ftCtx.RequestLogger)
		}
		groupMembers, err := sharing.FindMembersForGroup(ftCtx, groupID)
		if nil != err {
			return awsproxy.HandleError(err, ftCtx.RequestLogger)
		}
		var members []sharing.GroupMembers
		members = append(members, *groupMembers)
		return groupMembersResponse(ftCtx, members)
	}
	groups, err := sharing.FindGroupsForUser(ftCtx)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger)
	}
	var members []sharing.GroupMembers
	for group := 0; group < len(groups); group++ {
		groupID := groups[group]
		ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Building group %s", groupID))
		groupMembers, err := sharing.FindMembersForGroup(ftCtx, groupID)
		if nil != err {
			return awsproxy.HandleError(err, ftCtx.RequestLogger)
		}
		members = append(members, *groupMembers)
	}
	return groupMembersResponse(ftCtx, members)
}

func groupMembersResponse(ftCtx awsproxy.FTContext, groups []sharing.GroupMembers) awsproxy.Response {
	groupMembersResponse := GroupMembersResponse{
		PageToken: "page1",
		Groups:    groups,
	}
	return awsproxy.NewJSONResponse(ftCtx, groupMembersResponse)
}

func main() {
	lambda.Start(Handler)
}
