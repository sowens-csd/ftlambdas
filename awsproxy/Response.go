package awsproxy

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

const (
	accessOriginHeader      = "Access-Control-Allow-Origin"
	accessCredentialsHeader = "Access-Control-Allow-Credentials"
	contentTypeHeader       = "Content-Type"
	contentTypeJSON         = "application/json; charset=UTF-8"
	contentTypeText         = "text/plain; charset=UTF-8"
	originAll               = "*"
)

// FailureResponse can be used with the NewExpectedFailureResponse to send a structured JSON result
type FailureResponse struct {
	Code    int
	Message string
}

// NewJSONResponse creates a success response to return JSON in the body
func NewJSONResponse(ftCtx FTContext, dataForResponse interface{}) Response {
	body, err := json.Marshal(dataForResponse)
	if err != nil {
		return HandleError(err, ftCtx.RequestLogger)
	}
	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	bodyStr := buf.String()
	ftCtx.RequestLogger.Debug().Str("body", bodyStr).Msg("sending 200 ok, JSON response")
	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            bodyStr,
		Headers: map[string]string{
			contentTypeHeader:       contentTypeJSON,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	return resp
}

// NewResourceNotFoundResponse returns a standard 404 status response with the given body
func NewResourceNotFoundResponse(ftCtx FTContext, body string) Response {
	resp := Response{
		StatusCode:      404,
		IsBase64Encoded: false,
		Body:            body,
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// NewForbiddenResponse returns a standard 403 status response with the given body
func NewForbiddenResponse(ftCtx FTContext, body string) Response {
	resp := Response{
		StatusCode:      403,
		IsBase64Encoded: false,
		Body:            body,
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// NewForbiddenResponse returns a standard 403 status response with the given body
func NewUnauthorizedResponse(ftCtx FTContext) Response {
	resp := Response{
		StatusCode:      401,
		IsBase64Encoded: false,
		Body:            "Unauthorized",
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// NewSuccessResponse returns a standard 200 status response with a default body
func NewSuccessResponse(ftCtx FTContext) Response {
	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            "success",
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// NewTextResponse returns a standard 200 status response with the specified body
func NewTextResponse(ftCtx FTContext, body string) Response {
	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            body,
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// NewExpectedFailureResponse returns a standard 422 status response with the given body
func NewExpectedFailureResponse(ftCtx FTContext, dataForResponse interface{}) Response {
	body, err := json.Marshal(dataForResponse)
	if err != nil {
		return HandleError(err, ftCtx.RequestLogger)
	}
	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	resp := Response{
		StatusCode:      422,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			contentTypeHeader:       contentTypeJSON,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
	ftCtx.RequestLogger.Debug().Msg(fmt.Sprintf("Responding %d", resp.StatusCode))
	return resp
}

// HandleError creates a response for the error and logs it
func HandleError(err error, requestLogger zerolog.Logger) Response {
	requestLogger.Info().Err(err).Msg("error return")
	return Response{
		StatusCode:      500,
		IsBase64Encoded: false,
		Body:            err.Error(),
		Headers: map[string]string{
			contentTypeHeader:       contentTypeText,
			accessCredentialsHeader: "true",
			accessOriginHeader:      originAll,
		},
	}
}
