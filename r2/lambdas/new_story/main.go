package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	ftlambdas "github.com/sowens-csd/folktells-server"
	"github.com/sowens-csd/folktells-server/awsproxy"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	updateResult, err := ftlambdas.UpdateStoryAndNotify(ftCtx, request.Body, &http.Client{Timeout: 30 * time.Second})
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, updateResult), nil
}

func main() {
	lambda.Start(Handler)
}
