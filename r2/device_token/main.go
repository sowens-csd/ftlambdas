package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/notification"
)

// Handler is responsible for taking a request from a client mobile device that contains
// a new or changed device token and ensuring that the user record is updated to refer
// to that token.
//
// Device tokens identify a particular device for push notification. The token is stored
// so that it can be used in subsequent attempts to push a notification to that user
// and device.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	err := notification.RegisterFromRequest(ftCtx, request.Body)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
