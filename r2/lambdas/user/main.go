package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/sharing"
	"github.com/sowens-csd/folktells-server/store"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	return findUser(ftCtx), nil
}

func findUser(ftCtx awsproxy.FTContext) awsproxy.Response {
	onlineUser, err := sharing.LoadOnlineUser(ftCtx, ftCtx.UserID)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger)
	}
	if onlineUser.IsSubscriptionCheckRequired() {
		connectPassword, clientSecret, refreshToken, accessToken := awsproxy.AccessParameters()
		accessConfig := store.AccessConfig{
			ConnectPassword: connectPassword,
			ClientSecret:    clientSecret,
			RefreshToken:    refreshToken,
			AccessToken:     accessToken,
		}
		onlineUser = store.UpdateUserSubscription(ftCtx, onlineUser, accessConfig, &http.Client{Timeout: 30 * time.Second})
	}
	return awsproxy.NewJSONResponse(ftCtx, onlineUser)
}

func main() {
	lambda.Start(Handler)
}
