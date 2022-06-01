package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-cloud-go-lambda/awsproxy"
	"github.com/sowens-csd/folktells-cloud-go-lambda/ftauth"
)

func handler(ctx context.Context) error {
	ftCtx := awsproxy.NewFromContext(ctx, "n/a")
	return ftauth.CleanupOutstanding(ftCtx)
}

func main() {
	lambda.Start(handler)
}
