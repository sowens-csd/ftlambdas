package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/ftauth"
)

func handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	token := request.Headers["authorization"]
	tokenSlice := strings.Split(token, " ")
	var bearerToken string
	var email string
	if len(tokenSlice) > 2 {
		email = tokenSlice[len(tokenSlice)-2]
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}
	ftCtx := awsproxy.NewFromContext(ctx, "")
	if len(bearerToken) == 0 {
		ftCtx.RequestLogger.Debug().Str("token", token).Msg("No bearer token")
		return awsproxy.NewUnauthorizedResponse(ftCtx), nil
	}

	authResponse, err := ftauth.AuthenticateUser(ftCtx, email, bearerToken, false, time.Now)
	if nil != err {
		ftCtx.RequestLogger.Debug().Str("token", bearerToken).Msg("No token match")
		return awsproxy.NewUnauthorizedResponse(ftCtx), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, authResponse), nil
}

func main() {
	lambda.Start(handler)
}
