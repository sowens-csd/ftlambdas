package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/ftauth"
)

func handler(ctx context.Context, request events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx := awsproxy.NewFromContext(ctx, "")
	ftCtx.RequestLogger.Info().Str("version", request.Version).Str("type", request.Type).Msg("Request type")
	for headerName, headerVal := range request.Headers {
		ftCtx.RequestLogger.Info().Str("header", headerName).Str("Value", headerVal).Msg("headers")
	}
	token := request.Headers["authorization"]
	tokenSlice := strings.Split(token, " ")
	var bearerToken string
	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	} else {
		ftCtx.RequestLogger.Info().Str("token", token).Msg("Token missing or invalid format, expected 'bearer {token}'.")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}, errors.New("Unauthorized")
	}

	userID, email, err := ftauth.AuthorizeUser(ftCtx, bearerToken, time.Now)
	if nil != err {
		ftCtx.RequestLogger.Debug().Str("token", bearerToken).Err(err).Msg("No token match")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: false}, errors.New("Unauthorized")
	}
	allowedResource := os.Getenv("gatewayResources")
	ftCtx.RequestLogger.Debug().Str("email", email).Str("userID", userID).Str("res", allowedResource).Msg("Got user")
	return generatePolicy(email, userID, "Allow", allowedResource), nil
}

func generatePolicy(email, userID, effect, resource string) events.APIGatewayV2CustomAuthorizerSimpleResponse {
	authResponse := events.APIGatewayV2CustomAuthorizerSimpleResponse{IsAuthorized: true}

	authResponse.Context = map[string]interface{}{
		awsproxy.EmailContextField:  email,
		awsproxy.UserIDContextField: userID,
	}
	return authResponse
}

func main() {
	lambda.Start(handler)
}
