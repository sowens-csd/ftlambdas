package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftauth"
)

// Handler is responsible for taking a signup verify request from a client that
// verifies an add device request from another client.
//
// The flow is that the new device initiates an add device request, that is sent
// to existing devices via a push notification. An existing device can choose to
// allow or reject the request. If they allow it then the token for the login is
// added to verification request. The new device polls the system to see if the
// token is there, if it is it completes the add device cycle. If instead the
// signup request is gone then the request was denied.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	signupResp, err := ftauth.VerifyAddDevice(ftCtx, request.Body)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, signupResp), nil
}

func main() {
	lambda.Start(Handler)
}
