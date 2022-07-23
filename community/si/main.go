package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/ftdb"
)

type scheduledItem struct {
	ID             string     `json: "id",dynamodb:"id"`
	OrganizationID string     `json: "organizationID",dynamodb:"organizationID"`
	Name           string     `json: "name",dynamodb:"name"`
	Description    string     `json: "description,omitempty",dynamodb:"description,omitempty"`
	ScheduledAt    int        `json: "scheduledAt",dynamodb:"scheduledAt"`
	AllDay         bool       `json: "allDay",dynamodb:"allDay"`
	Tags           []ftdb.Tag `json: "tags",dynamodb:"tags"`
}

type scheduledItems struct {
	Items []scheduledItem `json: "items"`
}

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
	// // Prepare the S3 request so a signature can be generated

	// mediaCategory := request.PathParameters["mediaCategory"]
	// if mediaCategory != "user" {
	// 	ftCtx.RequestLogger.Info().Str("mediaCategory", mediaCategory).Msg("Unrecognized media category")
	// 	return awsproxy.HandleError(fmt.Errorf("Unrecognized media category %s", mediaCategory), ftCtx.RequestLogger), nil
	// }

	// mediaReferenceBytes, err := base64.URLEncoding.DecodeString(request.PathParameters["mediaReference"])
	// if err != nil {
	// 	ftCtx.RequestLogger.Info().Err(err).Msg("Failed to decode media file")
	// 	return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	// }
	// if len(contentType) == 0 {
	// 	return getMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	// } else {
	// 	return createMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	// }
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
	month, err := strconv.Atoi(monthParam)
	if nil != err {
		return awsproxy.HandleErrorV2(err, ftCtx.RequestLogger), nil
	}
	resID := ftdb.ResourceIDFromOrgAndMonth(orgID, month)
	itemRecords, err := ftdb.QueryByResource(ftCtx, resID)
	if nil != err {
		return awsproxy.HandleErrorV2(err, ftCtx.RequestLogger), nil
	}
	items := make([]scheduledItem, len(itemRecords))
	for i, itemRecord := range itemRecords {
		items[i] = scheduledItem{ID: itemRecord.ID, OrganizationID: itemRecord.OrganizationID, Name: itemRecord.Name, Description: itemRecord.Description, ScheduledAt: itemRecord.ScheduledAt, AllDay: itemRecord.AllDay, Tags: itemRecord.Tags}
	}
	// mediaReferenceBytes, err := base64.URLEncoding.DecodeString(request.PathParameters["mediaReference"])
	return awsproxy.NewJSONV2Response(ftCtx, scheduledItems{Items: items}), nil
}

func putScheduledItem(ftCtx awsproxy.FTContext, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	body := request.Body
	var item scheduledItem
	err := json.Unmarshal([]byte(body), &item)
	if nil != err {
		awsproxy.HandleErrorV2(err, ftCtx.RequestLogger)
	}
	refID := "SI#" + item.ID
	scheduledAt := time.UnixMilli(int64(item.ScheduledAt))
	resID := ftdb.ResourceIDFromOrgAndMonth(item.OrganizationID, int(scheduledAt.Month()))
	err = ftdb.PutItem(ftCtx, resID, refID, item)
	if nil != err {
		awsproxy.HandleErrorV2(err, ftCtx.RequestLogger)
	}
	return awsproxy.NewJSONV2Response(ftCtx, item), nil
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
