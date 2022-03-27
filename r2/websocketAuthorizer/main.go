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

func handler(ctx context.Context, request events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx := awsproxy.NewFromContext(ctx, "")
	token := request.Headers["Authorization"]
	tokenSlice := strings.Split(token, " ")
	var bearerToken string
	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	} else {
		ftCtx.RequestLogger.Info().Msg("Token missing or invalid format, expected 'bearer {token}'.")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	userID, email, err := ftauth.AuthorizeUser(ftCtx, bearerToken, time.Now)
	if nil != err {
		ftCtx.RequestLogger.Debug().Str("token", bearerToken).Err(err).Msg("No token match")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	allowedResource := os.Getenv("gatewayResources")
	ftCtx.RequestLogger.Debug().Str("email", email).Str("userID", userID).Str("res", allowedResource).Msg("Got user")
	return generatePolicy(email, userID, "Allow", allowedResource), nil
}

func generatePolicy(email, userID, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: userID}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
	authResponse.Context = map[string]interface{}{
		awsproxy.EmailContextField:  email,
		awsproxy.UserIDContextField: userID,
	}
	return authResponse
}

func main() {
	lambda.Start(handler)
}
