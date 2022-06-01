package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/sowens-csd/folktells-cloud-go-lambda/awsproxy"
	"github.com/sowens-csd/folktells-cloud-go-lambda/ftauth"
)

func handler(ctx context.Context, request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	awsproxy.SetupAccessParameters(ctx)
	token := request.AuthorizationToken
	tokenSlice := strings.Split(token, " ")
	var bearerToken string
	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}

	ftCtx := awsproxy.NewFromContext(ctx, "")
	userID, email, err := ftauth.AuthorizeUser(ftCtx, bearerToken)
	if nil != err {
		ftCtx.RequestLogger.WithFields(logrus.Fields{"token": bearerToken}).Debug("No token match")
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	allowedResource := os.Getenv("gatewayResources")
	ftCtx.RequestLogger.WithFields(logrus.Fields{"email": email, "userID": userID, "res": allowedResource}).Debug("Got user")
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
