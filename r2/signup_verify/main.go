package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftauth"
)

const requestIDParam = "requestID"

// Handler is responsible for taking a verification response for a signup from a client
// and verifying the token so that it can be used.
//
// When a user signs up they receive a link via email. The link points to this
// endpoint.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx := awsproxy.NewFromContext(ctx, "unknown")
	requestID := request.PathParameters[requestIDParam]
	verifyResp, err := ftauth.VerifySignup(ftCtx, requestID)
	if nil != err {
		return awsproxy.NewUnauthorizedResponse(ftCtx), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, verifyResp), nil
}

func main() {
	lambda.Start(Handler)
}
