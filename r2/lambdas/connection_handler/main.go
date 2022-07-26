package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/messaging"
	"github.com/sowens-csd/folktells-server/notification"
)

func main() {
	lambda.Start(handler)
}

// handler receives a synchronous invocation from API Gateway when a WebSocket connection is opened for the
// application's API.
func handler(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (awsproxy.Response, error) {

	ftCtx, errResp := awsproxy.NewFromWebsocketContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	err := notification.InitHost(ftCtx, request.RequestContext.DomainName)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	ftCtx.RequestLogger.Debug().Str("user", ftCtx.UserID).Str("domain", request.RequestContext.DomainName).Str("connection", request.RequestContext.ConnectionID).Msg("Connection handler called")
	err = messaging.RecordConnection(ftCtx, request.RequestContext.ConnectionID)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewSuccessResponse(ftCtx), nil
}
