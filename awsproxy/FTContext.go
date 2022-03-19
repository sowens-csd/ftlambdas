package awsproxy

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// EmailContextField The key for the field that holds the user's email address in the
	// Authorizer context
	EmailContextField = "ftEmail"
	// UserIDContextField The key for the field that holds the user's ID in the
	// Authorizer context
	UserIDContextField = "ftUserId"
	requestIDLogField  = "request_id"
	userIDLogField     = "user_id"
)

const testRequestID string = "request1"

type FTDynamoAPI interface {
	PutItem(ctx context.Context, input *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, input *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// FTContext information needed by methods in the call chain
type FTContext struct {
	Context       context.Context
	RequestID     string
	UserID        string
	Email         string
	Config        config.Config
	DBSvc         FTDynamoAPI
	RequestLogger zerolog.Logger
	EmailSvc      emailSender
}

// NewFromContext create a new FTContext from an inbound proxy context
func NewFromContext(ctx context.Context, userID string) FTContext {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	requestID := lambdaContext.AwsRequestID
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}
	svc := dynamodb.NewFromConfig(cfg)
	setLogLevel()
	requestLogger := log.With().Str(requestIDLogField, requestID).Logger()
	return FTContext{
		Context:       ctx,
		RequestID:     requestID,
		UserID:        userID,
		Config:        cfg,
		DBSvc:         svc,
		RequestLogger: requestLogger,
		EmailSvc:      &sesEmailSender{},
	}
}

// NewFromContextAndClaims create a new FTContext from an inbound proxy context
func NewFromContextAndClaims(ctx context.Context, request Request) (FTContext, *Response) {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	requestID := lambdaContext.AwsRequestID
	claims := request.RequestContext.Authorizer["claims"]
	if nil == claims {
		requestLogger := log.With().Str(requestIDLogField, requestID).Logger()
		requestLogger.Error().Msg("claims null")
		errResp := HandleError(fmt.Errorf("No claims in the request"), requestLogger)
		return FTContext{}, &errResp
	}
	userID := claims.(map[string]interface{})["username"].(string)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}
	svc := dynamodb.NewFromConfig(cfg)
	setLogLevel()
	requestLogger := log.With().Str(requestIDLogField, requestID).Str(userIDLogField, userID).Logger()
	return FTContext{
		Context:       ctx,
		RequestID:     requestID,
		UserID:        userID,
		DBSvc:         svc,
		RequestLogger: requestLogger,
		EmailSvc:      &sesEmailSender{},
	}, nil
}

// NewFromContextAndJWT create a new FTContext from an inbound proxy context
// that was authorized through the custom JWT authorizer.
//
// The JWT authorizer adds new claims with the internal Folktells user id and
// email address to the request.
func NewFromContextAndJWT(ctx context.Context, request Request) (FTContext, *Response) {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	requestID := lambdaContext.AwsRequestID
	claims := request.RequestContext.Authorizer
	if nil == claims {
		requestLogger := log.With().Str("request_id", requestID).Logger()
		requestLogger.Error().Msg("claims null: ")
		// requestLogger.Error().Msg(request.RequestContext.Authorizer)
		errResp := HandleError(fmt.Errorf("No claims in the request"), requestLogger)
		return FTContext{}, &errResp
	}
	userID := claims[UserIDContextField].(string)
	email := claims[EmailContextField].(string)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}
	svc := dynamodb.NewFromConfig(cfg)
	setLogLevel()
	requestLogger := log.With().Str(requestIDLogField, requestID).Str(userIDLogField, userID).Logger()
	return FTContext{
		Context:       ctx,
		RequestID:     requestID,
		UserID:        userID,
		Email:         email,
		DBSvc:         svc,
		RequestLogger: requestLogger,
		EmailSvc:      &sesEmailSender{},
	}, nil
}

// NewFromContextAndJWT create a new FTContext from an inbound proxy context
// that was authorized through the custom JWT authorizer.
//
// The JWT authorizer adds new claims with the internal Folktells user id and
// email address to the request.
func NewFromWebsocketContextAndJWT(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (FTContext, *Response) {
	lambdaContext, _ := lambdacontext.FromContext(ctx)
	requestID := lambdaContext.AwsRequestID
	claims := request.RequestContext.Authorizer.(map[string]interface{})
	if nil == claims {
		requestLogger := log.With().Str("request_id", requestID).Logger()
		requestLogger.Error().Msg("claims null: ")
		// requestLogger.Error().Msg(request.RequestContext.Authorizer)
		errResp := HandleError(fmt.Errorf("No claims in the request"), requestLogger)
		return FTContext{}, &errResp
	}
	userID := claims[UserIDContextField].(string)
	email := claims[EmailContextField].(string)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("Could not load AWS config")
	}
	svc := dynamodb.NewFromConfig(cfg)
	setLogLevel()
	requestLogger := log.With().Str(requestIDLogField, requestID).Str(userIDLogField, userID).Logger()
	return FTContext{
		Context:       ctx,
		RequestID:     requestID,
		UserID:        userID,
		Email:         email,
		DBSvc:         svc,
		RequestLogger: requestLogger,
		EmailSvc:      &sesEmailSender{},
	}, nil
}

// NewTestContext used for test cases
func NewTestContext(userID string, db FTDynamoAPI) FTContext {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	requestLogger := log.With().Str(requestIDLogField, testRequestID).Str(userIDLogField, userID).Logger()
	return FTContext{
		RequestID:     testRequestID,
		UserID:        userID,
		DBSvc:         db,
		RequestLogger: requestLogger,
		EmailSvc:      &TestEmailSender{},
	}
}

func setLogLevel() {
	levelName := os.Getenv("LOG_LEVEL")
	logLevel := zerolog.InfoLevel
	switch levelName {
	case "debug":
		logLevel = zerolog.DebugLevel
		break
	case "error":
		logLevel = zerolog.ErrorLevel
		break
	case "trace":
		logLevel = zerolog.TraceLevel
		break
	}
	zerolog.SetGlobalLevel(logLevel)
}
