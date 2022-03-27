package main

import (
	"context"
	"net/url"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/webrtc"
)

// Handler takes a deviceId in the body of the request and uses it create or find a
// ChannelArn which it then returns in the body of the response.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	viewer := request.PathParameters["vm"] == "v"
	encodedChannelARN := request.PathParameters["channel"]
	channelARN, err := url.PathUnescape(encodedChannelARN)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	services, err := webrtc.GetServices(ftCtx, channelARN, request.PathParameters["device"], viewer)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, services), nil
}

func main() {
	lambda.Start(Handler)
}
