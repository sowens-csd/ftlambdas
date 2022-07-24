package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/si"
)

// Handler for all requests to the various Scheduled Item endpoints, these can variously:
// - List items for an organization in a given month
// - Create a new item
func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	ftCtx, errResp := awsproxy.NewFromV2ContextAndJWT(ctx, request)
	if nil != errResp {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusForbidden, Body: "Forbidden"}, nil
	}
	ftCtx.RequestLogger.Debug().Str("userID", ftCtx.UserID).Msg("scheduled item handler")

	httpRequest := request.RequestContext.HTTP
	if httpRequest.Method == "GET" {
		return getScheduledItems(ftCtx, request)
	} else if httpRequest.Method == "POST" {
		return putScheduledItem(ftCtx, request)
	}
	return events.APIGatewayProxyResponse{StatusCode: http.StatusNotFound, Body: fmt.Sprintf("Path: %s, Method: %s", request.RequestContext.HTTP.Path, request.RequestContext.HTTP.Method)}, nil
}

func getScheduledItems(ftCtx awsproxy.FTContext, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	orgID, err := getParam(request, "org")
	if nil != err {
		return awsproxy.HandleErrorV2(err, ftCtx.RequestLogger), nil
	}
	monthParam, ok := request.PathParameters["month"]
	if !ok {
		return awsproxy.HandleErrorV2(fmt.Errorf("month path parameter missing"), ftCtx.RequestLogger), nil
	}
	scheduledItems, err := si.GetScheduledItems(ftCtx, orgID, monthParam)
	if nil != err {
		return awsproxy.HandleErrorV2(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONV2Response(ftCtx, scheduledItems), nil
}

func putScheduledItem(ftCtx awsproxy.FTContext, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	scheduledItem, err := si.PutScheduledItem(ftCtx, request.Body)
	if nil != err {
		awsproxy.HandleErrorV2(err, ftCtx.RequestLogger)
	}
	return awsproxy.NewJSONV2Response(ftCtx, scheduledItem), nil
}

func getParam(request events.APIGatewayV2HTTPRequest, paramName string) (string, error) {
	paramBytes, err := base64.URLEncoding.DecodeString(request.PathParameters[paramName])
	if nil != err {
		return "", err
	}
	return string(paramBytes), nil
}

func main() {
	lambda.Start(Handler)
}
