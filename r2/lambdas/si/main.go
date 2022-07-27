package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/si"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	orgID, err := getParam(request, "orgId")
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	monthParam, ok := request.PathParameters["month"]
	if !ok {
		return awsproxy.HandleError(fmt.Errorf("month path parameter missing"), ftCtx.RequestLogger), nil
	}
	yearParam, ok := request.PathParameters["year"]
	if !ok {
		return awsproxy.HandleError(fmt.Errorf("year path parameter missing"), ftCtx.RequestLogger), nil
	}
	ftCtx.RequestLogger.Debug().Str("orgID", orgID).Str("month", monthParam).Str("year", yearParam).Msg("get scheduled items by")
	scheduledItems, err := si.GetScheduledItems(ftCtx, orgID, yearParam, monthParam)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, scheduledItems), nil
}

func main() {
	lambda.Start(Handler)
}

func getParam(request awsproxy.Request, paramName string) (string, error) {
	paramBytes, err := base64.URLEncoding.DecodeString(request.PathParameters[paramName])
	if nil != err {
		return "", err
	}
	return string(paramBytes), nil
}
