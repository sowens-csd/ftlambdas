package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/ftlambdas/awsproxy"
	"github.com/sowens-csd/ftlambdas/notification"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	responseBody := `<?xml version="1.0" encoding="UTF-8"?>
	<Response><Message><Body>Hello world! -Lambda</Body></Message></Response>
	`
	ftCtx := awsproxy.NewFromContext(ctx, "unknown")
	err := notification.SendFromSMS(ftCtx, request.Body, &http.Client{Timeout: 30 * time.Second})
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            responseBody,
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
