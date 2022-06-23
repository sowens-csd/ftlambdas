package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/ftauth"
)

// Handler is responsible for taking a signup request from a client that contains
// a new or changed user token and confirming the request.
//
// Client devices generate a token then ask to have the token confirmed. Confirmation
// is done by sending a link to the incldued email address. If the user can receive
// the link and click it then the token is confirmed.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx := awsproxy.NewFromContext(ctx, "unknown")
	signupResp, err := ftauth.Signup(ftCtx, request.Body, &http.Client{Timeout: 30 * time.Second})
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, signupResp), nil
}

func main() {
	lambda.Start(Handler)
}
