package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
	"github.com/sowens-csd/folktells-cloud-go-lambda/awsproxy"
	"github.com/sowens-csd/folktells-cloud-go-lambda/ftdb"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	var req ftdb.QueryRequest
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return awsproxy.HandleError(fmt.Errorf("Bad request"), ftCtx.RequestLogger), nil
	}
	ftCtx.RequestLogger.WithFields(logrus.Fields{"resID": req.ResourceID, "refID": req.ReferenceID}).Debug("Deleting")
	err = ftdb.DeleteItem(ftCtx, req.ResourceID, req.ReferenceID)
	if nil != err {
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
