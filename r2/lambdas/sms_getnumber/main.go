package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/sms"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	country, found := request.QueryStringParameters["isoCountry"]
	if len(country) == 0 || !found {
		return awsproxy.HandleError(fmt.Errorf("isoCountry is required"), ftCtx.RequestLogger), nil
	}
	areaCode := request.QueryStringParameters["areaCode"]

	availableNumbers, err := sms.GetAvailableNumbers(ftCtx, country, areaCode, &http.Client{Timeout: 30 * time.Second})
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}

	return awsproxy.NewJSONResponse(ftCtx, availableNumbers), nil
}

func main() {
	lambda.Start(Handler)
}
