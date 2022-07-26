package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/app"
	"github.com/sowens-csd/folktells-server/awsproxy"
)

type versionCheckResponse struct {
	Status int `json:"status"`
}

// Handler takes either a JSON analytics event and saves it to the analytics
// stream, or an appVersion paramter and validates if that version is supported
// by the back-end.
// https://devapi.folktells.com/r2/app/usage?appVersion=5.3.8
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx := awsproxy.NewFromContext(ctx, "")
	appVersion := request.QueryStringParameters["appVersion"]
	if len(appVersion) > 0 {
		incomingVersion := appVersion
		subVersionPairs := strings.Split(appVersion, ".")
		if len(subVersionPairs) > 1 {
			appVersion = fmt.Sprintf("%s.%s", subVersionPairs[0], subVersionPairs[1])
		}
		ftCtx.RequestLogger.Debug().Str("appVersion", appVersion).Msg("version check")
		switch appVersion {
		case "5.3":
			fallthrough
		case "5.4":
			fallthrough
		case "5.5":
			fallthrough
		case "5.6":
			return awsproxy.NewJSONResponse(ftCtx, versionCheckResponse{Status: 0}), nil
		}
		ftCtx.RequestLogger.Info().Str("appVersion", incomingVersion).Str("sub", appVersion).Msg("failed version check")
		return awsproxy.NewJSONResponse(ftCtx, versionCheckResponse{Status: 1}), nil
	} else {
		// ftCtx.RequestLogger.Info().Str("sub", appVersion).Msg("no app version provided")
		err := app.SaveAnalyticEvents(ftCtx, request.Body)
		if nil != err {
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		return awsproxy.NewSuccessResponse(ftCtx), nil
	}
}

func main() {
	lambda.Start(Handler)
}
