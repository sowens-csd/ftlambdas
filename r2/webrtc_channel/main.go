package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/webrtc"
)

// Handler takes a deviceId in the body of the request and uses it create or find a
// ChannelArn which it then returns in the body of the response.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	channelArn, err := webrtc.CreateChannel(ftCtx, request.Body)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewTextResponse(ftCtx, channelArn), nil
}

func main() {
	lambda.Start(Handler)
}
