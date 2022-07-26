package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
	"github.com/sowens-csd/folktells-server/sharing"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	awsproxy.SetupAccessParameters(ctx)
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	err := sharing.AssignNumber(ftCtx, request.Body, &http.Client{Timeout: 30 * time.Second})
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}

	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
