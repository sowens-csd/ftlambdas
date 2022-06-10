package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sharing"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	return updateGroupMember(ftCtx, request.Body), nil
}

// Note that the memberJSON could be either a GroupMember or a GroupMemberInvitation
func updateGroupMember(ftCtx awsproxy.FTContext, memberJSON string) awsproxy.Response {
	err := sharing.UpdateGroupMembership(ftCtx, memberJSON)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger)
	}
	return awsproxy.NewSuccessResponse(ftCtx)
}

func main() {
	lambda.Start(Handler)
}
