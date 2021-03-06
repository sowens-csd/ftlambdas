package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-cloud-go-lambda/awsproxy"
	"github.com/sowens-csd/folktells-cloud-go-lambda/ftdb"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	result, err := ftdb.QueryFromRequest(ftCtx, request.Body)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewJSONResponse(ftCtx, result), nil
}

func main() {
	lambda.Start(Handler)
}
