package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	folktells "github.com/sowens-csd/folktells-server"
	"github.com/sowens-csd/folktells-server/awsproxy"
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
	ftCtx.RequestLogger.Info().Msg("About to AddManagedUser")
	managedUser, err := folktells.AddManagedUser(ftCtx, string(request.Body))
	if nil != err {
		ftCtx.RequestLogger.Info().Err(err).Msg("Error Adding ManagedUser")
		return awsproxy.NewTextResponse(ftCtx, "failed"), nil
	}
	ftCtx.RequestLogger.Info().Msg("After Adding ManagedUser")
	return awsproxy.NewJSONResponse(ftCtx, *managedUser), nil
}

func main() {
	fmt.Print("Starting FolkCreate")
	lambda.Start(Handler)
	fmt.Print("Started FolkCreate")
}
