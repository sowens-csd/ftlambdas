package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sowens-csd/folktells-server/awsproxy"
)

// Handler for all requests to the various tag endpoints, these can variously:
// - List all existing tags
// - List tags for a specific resource
// - Tag a specific resource
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	ftCtx.RequestLogger.Info().Msg("About to create media access URL")

	// // Prepare the S3 request so a signature can be generated

	// mediaCategory := request.PathParameters["mediaCategory"]
	// if mediaCategory != "user" {
	// 	ftCtx.RequestLogger.Info().Str("mediaCategory", mediaCategory).Msg("Unrecognized media category")
	// 	return awsproxy.HandleError(fmt.Errorf("Unrecognized media category %s", mediaCategory), ftCtx.RequestLogger), nil
	// }

	// mediaReferenceBytes, err := base64.URLEncoding.DecodeString(request.PathParameters["mediaReference"])
	// if err != nil {
	// 	ftCtx.RequestLogger.Info().Err(err).Msg("Failed to decode media file")
	// 	return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	// }
	// if len(contentType) == 0 {
	// 	return getMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	// } else {
	// 	return createMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	// }
	return awsproxy.NewSuccessResponse(ftCtx), nil
}

func main() {
	lambda.Start(Handler)
}
