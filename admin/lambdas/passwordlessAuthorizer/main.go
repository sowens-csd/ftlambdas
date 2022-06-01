package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/sowens-csd/folktells-cloud-go-lambda/awsproxy"
	"github.com/sowens-csd/folktells-cloud-go-lambda/ftauth"
)

func handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	token := request.Headers["Authorization"]
	ftCtx := awsproxy.NewFromContext(ctx, "")
	authResponse, err := ftauth.AuthenticateUserWithToken(ftCtx, token, true)
	if nil != err {
		ftCtx.RequestLogger.WithFields(logrus.Fields{"token": token}).Debug("No token match")
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, authResponse), nil
}

func main() {
	lambda.Start(handler)
}
