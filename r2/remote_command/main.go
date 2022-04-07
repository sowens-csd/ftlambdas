package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/notification"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	err := notification.InitHost(ftCtx, "")
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	err = notification.SendFromCommand(ftCtx, request.Body)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
