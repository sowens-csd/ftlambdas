package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/sharing"
)

type Organization struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type organizationList struct {
	Count  int            `json:"count"`
	Result []Organization `json:"result"`
}

type folkList struct {
	Count  int                  `json:"count"`
	Result []sharing.OnlineUser `json:"result"`
}

const orgIDParam = "orgID"
const subtypeParam = "subtype"

// Handler is responsible for taking one of the possible org requests and
// producing the desired result.
//
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	subtype, hasSubtype := request.PathParameters[subtypeParam]
	if orgID, ok := request.PathParameters[orgIDParam]; ok {
		if hasSubtype && subtype == "folk" {
			ftCtx.RequestLogger.Debug().Str("orgId", orgID).Msg("About to list folk in org")
			managedUsers, err := sharing.FindManagedUsers(ftCtx, orgID)
			if nil != err {
				ftCtx.RequestLogger.Info().Str("orgID", orgID).Err(err).Msg("Error finding users")
				return awsproxy.NewTextResponse(ftCtx, "failed"), nil
			}
			return awsproxy.NewJSONResponse(ftCtx, folkList{Count: len(managedUsers), Result: managedUsers}), nil
		}
	} else {
		ftCtx.RequestLogger.Debug().Msg("About to list orgs")
		resultArr := make([]Organization, 0)
		resultArr = append(resultArr, Organization{Name: "Oakpark", ID: "64e5d97b-ca28-452f-87df-79d1eab1ad7e"})
		return awsproxy.NewJSONResponse(ftCtx, organizationList{Count: len(resultArr), Result: resultArr}), nil
	}
	ftCtx.RequestLogger.Debug().Str("path", request.Path).Str("subtype", subtype).Msg("Did not recognize the request path")
	return awsproxy.NewResourceNotFoundResponse(ftCtx, "Path not recognized"), nil
}

func main() {
	lambda.Start(Handler)
}
