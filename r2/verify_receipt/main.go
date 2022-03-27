package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/store"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	connectPassword, clientSecret, refreshToken, accessToken := awsproxy.AccessParameters()
	accessConfig := store.AccessConfig{
		ConnectPassword: connectPassword,
		ClientSecret:    clientSecret,
		RefreshToken:    refreshToken,
		AccessToken:     accessToken,
	}
	verifyResp, updatedToken := store.ProcessPurchaseRequest(ftCtx, request.Body, accessConfig, &http.Client{Timeout: 30 * time.Second})
	awsproxy.UpdateAccessToken(ctx, updatedToken)
	return awsproxy.NewJSONResponse(ftCtx, verifyResp), nil
}

func main() {
	lambda.Start(Handler)
}
