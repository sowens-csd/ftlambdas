package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/sharing"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	groupIDEnc, found := request.PathParameters["groupID"]
	if !found {
		return awsproxy.HandleError(fmt.Errorf("GroupID path parameter missing."), ftCtx.RequestLogger), nil
	}
	groupID, err := url.QueryUnescape(groupIDEnc)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	group, err := sharing.LoadGroup(ftCtx, groupID)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	if group.OwnerID != ftCtx.UserID {
		return awsproxy.NewForbiddenResponse(ftCtx, "Only the group owner can delete."), nil
	}
	err = group.Delete(ftCtx)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
