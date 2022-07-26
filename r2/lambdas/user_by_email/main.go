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
	emailEnc, found := request.PathParameters["email"]
	if found {
		email, err := url.QueryUnescape(emailEnc)
		if err != nil {
			return awsproxy.HandleError(fmt.Errorf("Bad email encoding"), ftCtx.RequestLogger), nil
		}
		return findUser(ftCtx, email), nil
	}
	return awsproxy.HandleError(fmt.Errorf("Email path parameter missing"), ftCtx.RequestLogger), nil
}

func findUser(ftCtx awsproxy.FTContext, email string) awsproxy.Response {
	onlineUser, err := sharing.LoadOnlineUserByEmail(ftCtx, email)
	if nil == err && nil != onlineUser {
		return awsproxy.NewJSONResponse(ftCtx, onlineUser)
	} else {
		switch err.(type) {
		case *sharing.UserNotFoundError:
			return awsproxy.NewResourceNotFoundResponse(ftCtx, fmt.Sprintf("No user with email %s", email))
			break
		}
	}
	return awsproxy.NewResourceNotFoundResponse(ftCtx, fmt.Sprintf("No user with email %s", email))
}

func main() {
	lambda.Start(Handler)
}
