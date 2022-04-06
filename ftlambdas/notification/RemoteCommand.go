package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/messaging"

	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	apigtypes "github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi/types"
)

// cfg is the base or parent AWS configuration for this lambda.
var cfg aws.Config

// apiClient provides access to the Amazon API Gateway management functions. Once initialized, the instance is reused
// across subsequent AWS Lambda invocations. This potentially amortizes the instance creation over multiple executions
// of the AWS Lambda instance.
var apiClient *apigatewaymanagementapi.Client

// SendCommand sends the provided command to all users in the same group.
func SendCommand(ftCtx awsproxy.FTContext, details commandDetails) error {

	ftCtx.RequestLogger.Debug().Str("toUser", details.ToUser).Msg("send command to users")
	connectionIds, err := messaging.LoadUserConnections(ftCtx, details.ToUser)
	if nil != err {
		return err
	}
	ftCtx.RequestLogger.Debug().Int("connections", len(connectionIds)).Msg("found users")
	commandJSON := new(bytes.Buffer)
	encoder := json.NewEncoder(commandJSON)
	encoder.Encode(&details)
	ftCtx.RequestLogger.Debug().Msg("message encoded")
	for _, id := range connectionIds {
		err := publish(ftCtx, id, commandJSON.Bytes())
		ftCtx.RequestLogger.Debug().Str("connectionId", id).Msg("sent")
		if nil != err {
			err = handleError(ftCtx, err, details.ToUser, id)
			if nil != err {
				ftCtx.RequestLogger.Info().Err(err).Str("connectionId", id).Msg("problem handling error")
			}
		}
	}

	return nil
}

// publish publishes the provided data to the provided Amazon API Gateway connection ID. A common failure scenario which
// results in an error is if the connection ID is no longer valid. This can occur when a client disconnected from the
// Amazon API Gateway endpoint but the disconnect AWS Lambda was not invoked as it is not guaranteed to be invoked when
// clients disconnect.
func publish(ftCtx awsproxy.FTContext, id string, data []byte) error {
	initConfig(ftCtx)
	ftCtx.RequestLogger.Debug().Str("connectionId", id).Int("bytes to send", len(data)).Bool("context is nil", nil == ftCtx.Context).Msg("publishing message")
	ctx, _ := context.WithTimeout(ftCtx.Context, 10*time.Second)
	_, err := apiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		Data:         data,
		ConnectionId: &id,
	})

	return err
}

func initConfig(ftCtx awsproxy.FTContext) {
	if nil == apiClient {
		var err error
		cfg, err = config.LoadDefaultConfig(ftCtx.Context)
		if err != nil {
			panic("unable to load SDK config")
		}
		apiClient = newAPIGatewayManagementClient(ftCtx, &cfg, ftCtx.DomainName, ftCtx.Stage)
		ftCtx.RequestLogger.Debug().Str("domain", ftCtx.DomainName).Str("stage", ftCtx.Stage).Msg("created client")
	}
}

// newAPIGatewayManagementClient creates a new API Gateway Management Client instance from the provided parameters. The
// new client will have a custom endpoint that resolves to the application's deployed API.
func newAPIGatewayManagementClient(ftCtx awsproxy.FTContext, cfg *aws.Config, domain, stage string) *apigatewaymanagementapi.Client {
	cp := cfg.Copy()
	cp.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		ftCtx.RequestLogger.Debug().Str("service", service).Str("region", region).Msg("resolving endpoint")
		// if service != "execute-api" {
		// 	ftCtx.RequestLogger.Debug().Bool("nilCfg", nil == cfg).Bool("nilResolver", nil == cfg.EndpointResolver).Msg("default endpoint resolver")
		// 	return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		// }

		var endpoint url.URL
		endpoint.Path = stage
		endpoint.Host = "7m9oa3fcn6.execute-api.ca-central-1.amazonaws.com"
		endpoint.Scheme = "https"
		ftCtx.RequestLogger.Debug().Str("url", endpoint.String()).Msg("resolved endpoint")
		return aws.Endpoint{
			SigningRegion: region,
			URL:           endpoint.String(),
		}, nil
	})

	return apigatewaymanagementapi.NewFromConfig(cp)
}

// handleError is a convenience function for taking action for a given error value. The function handles nil errors as a
// convenience to the caller. If a nil error is provided, the error is immediately returned. The function may return an
// error from the handling action, such as deleting the id from the cache, if that action results in an error.
func handleError(ftCtx awsproxy.FTContext, err error, userID, id string) error {
	if err == nil {
		return err
	}

	switch err.(type) {
	case *apigtypes.GoneException:
		ftCtx.RequestLogger.Info().Err(err).Str("connectionId", id).Msg("gone - delete stale connection details from cache")
		return deleteConnectionId(ftCtx, userID, id)
	default:
		ftCtx.RequestLogger.Info().Err(err).Str("connectionId", id).Msg("unk - delete connection details from cache")
		return deleteConnectionId(ftCtx, userID, id)
	}
}

func deleteConnectionId(ftCtx awsproxy.FTContext, userID, id string) error {
	initConfig(ftCtx)
	messaging.RemoveConnection(ftCtx, userID, id)
	_, err := apiClient.DeleteConnection(ftCtx.Context, &apigatewaymanagementapi.DeleteConnectionInput{ConnectionId: &id})

	return err
}
