package awsproxy

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rs/zerolog/log"
)

// Request is of type APIGatewayProxyRequest since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
type Request events.APIGatewayProxyRequest

// GetRequestAndUser returns the unique requestID for the proxy request and the Cognito userID making it
func GetRequestAndUser(ctx context.Context, request Request) (string, string, *Response) {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	requestID := lambdaContext.AwsRequestID
	claims := request.RequestContext.Authorizer["claims"]
	log.Logger.Debug().Msg("After claims retrieve")
	if nil == claims {
		requestLogger := log.With().Str("request_id", requestID).Logger()
		requestLogger.Error().Msg("claims null")
		errResp := HandleError(fmt.Errorf("No claims in the request"), requestLogger)
		return "", "", &errResp
	}
	log.Debug().Msg(fmt.Sprint(claims))
	userID := claims.(map[string]interface{})["username"].(string)
	return requestID, userID, nil
}
