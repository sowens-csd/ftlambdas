package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	signer "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/sowens-csd/folktells-server/awsproxy"
)

const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type mediaAccessResponse struct {
	PutURL string `json:"putURL"`
	GetURL string `json:"getURL"`
}

// Handler is responsible for taking a signup verify request from a client that
// verifies an add device request from another client.
//
// The flow is that the new device initiates an add device request, that is sent
// to existing devices via a push notification. An existing device can choose to
// allow or reject the request. If they allow it then the token for the login is
// added to verification request. The new device polls the system to see if the
// token is there, if it is it completes the add device cycle. If instead the
// signup request is gone then the request was denied.
func Handler(ctx context.Context, request awsproxy.Request) (awsproxy.Response, error) {
	ftCtx, errResp := awsproxy.NewFromContextAndJWT(ctx, request)
	if nil != errResp {
		return *errResp, nil
	}
	ftCtx.RequestLogger.Info().Msg("About to create media access URL")
	// Load the bucket name
	s3Bucket := os.Getenv("s3Bucket")
	if s3Bucket == "" {
		log.Fatal("an s3 bucket was unable to be loaded from env vars")
	}

	// Prepare the S3 request so a signature can be generated

	accessToken, secretKey := awsproxy.SharedCredentialParameters(ftCtx.Context)
	os.Setenv("AWS_ACCESS_KEY_ID", accessToken)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	os.Setenv("AWS_REGION", "ca-central-1")
	os.Setenv("AWS_DEFAULT_REGION", "ca-central-1")
	// customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
	// 	return aws.Endpoint{
	// 		PartitionID:   "aws",
	// 		URL:           "https://s3.ca-central-1.amazonaws.com",
	// 		SigningRegion: "ca-central-1",
	// 	}, nil
	// })
	// cfg, err := config.LoadDefaultConfig(ftCtx.Context, config.WithEndpointResolverWithOptions(customResolver))
	// cfg, err := config.LoadDefaultConfig(ftCtx.Context)
	// cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKID", "SECRET_KEY", "TOKEN")))
	// if err != nil {
	// 	ftCtx.RequestLogger.Info().Err(err).Msg("Failed to load config")
	// 	return awsproxy.NewForbiddenResponse(ftCtx, "media access failed"), nil
	// }

	// svc := s3.NewFromConfig(cfg)
	// presigner := s3.NewPresignClient(svc)
	mediaFile := request.PathParameters["mediaFile"]

	ftCtx.RequestLogger.Debug().Int("AccessKey", len(accessToken)).Int("SecretKey", len(secretKey)).Str("mediaFile", mediaFile).Str("bucket", s3Bucket).Msg("Have params for presign")
	// putReq, err := presigner.PresignPutObject(ftCtx.Context, &s3.PutObjectInput{Bucket: &s3Bucket, Key: &mediaFile}, s3.WithPresignExpires(1*time.Hour))
	// if err != nil {
	// 	ftCtx.RequestLogger.Info().Err(err).Msg("Failed to generate pre-signed PUT url")
	// 	return awsproxy.NewForbiddenResponse(ftCtx, "media access failed"), nil
	// }
	// getReq, err := presigner.PresignGetObject(ftCtx.Context, &s3.GetObjectInput{Bucket: &s3Bucket, Key: &mediaFile}, s3.WithPresignExpires(1*time.Hour))
	// if err != nil {
	// 	ftCtx.RequestLogger.Info().Err(err).Msg("Failed to generate pre-signed GET url")
	// 	return awsproxy.NewForbiddenResponse(ftCtx, "media access failed"), nil
	// }
	putURL, getURL, err := generateSigned(ftCtx, accessToken, secretKey, mediaFile, s3Bucket)
	if err != nil {
		ftCtx.RequestLogger.Info().Err(err).Msg("Failed to manually generate signed URL")
	}
	var response = mediaAccessResponse{
		PutURL: putURL,
		GetURL: getURL,
	}

	// ftCtx.RequestLogger.Debug().Str("put", putReq.URL).Int("putLen", len(putReq.URL)).Str("get", getReq.URL).Int("getLen", len(getReq.URL)).Str("man-put", putURL).Str("man-get", getURL).Msg("Generated pre-signed urls")
	ftCtx.RequestLogger.Debug().Str("man-put", putURL).Str("man-get", getURL).Msg("Generated pre-signed urls")
	return awsproxy.NewJSONResponse(ftCtx, response), nil
	// return awsproxy.NewTextResponse(ftCtx, fmt.Sprintf("PUT:\n%s\n\nGET:\n%s\n\nManual PUT:\n%s\n\nManual GET:\n%s", putReq.URL, getReq.URL, putURL, getURL)), nil
}

func generateSigned(ftCtx awsproxy.FTContext, accessKeyID, secretKey, mediaFile, s3Bucket string) (string, string, error) {
	httpSigner := signer.NewSigner()
	uri := fmt.Sprintf("https://sharedstack-folktellsmediabucket60d66dfa-umnguk71tkci.s3.ca-central-1.amazonaws.com/%s", url.PathEscape(mediaFile))
	req, _ := http.NewRequest("GET", uri, nil)
	params := url.Values{
		"X-Amz-Credential": {fmt.Sprintf("%s/%s/ca-central-1/s3/aws4_request", accessKeyID, time.Now().Format("20060102"))},
		"X-Amz-Expires":    {"3600"},
		"x-id":             {"GetObject"},
		"x-amz-acl":        {"public-read"},
		"X-Amz-Algorithm":  {"AWS4-HMAC-SHA256"},
	}
	// req.Header.Set("Content-Type", "application/jpg")
	req.URL.RawQuery = params.Encode()
	appCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyID, secretKey, ""))
	creds, err := appCreds.Retrieve(ftCtx.Context)
	if err != nil {
		return "", "", err
	}

	// PresignHTTP(ctx context.Context, credentials aws.Credentials, r *http.Request,payloadHash string, service string, region string, signingTime time.Time,optFns ...func(*SignerOptions),) (signedURI string, signedHeaders http.Header, err error)
	getURL, _, err := httpSigner.PresignHTTP(ftCtx.Context, creds, req, "UNSIGNED-PAYLOAD", "s3", "ca-central-1", time.Now())
	if err != nil {
		return "", "", err
	}
	req, _ = http.NewRequest("PUT", uri, nil)
	params = url.Values{
		"key":              {fmt.Sprintf("/%s", mediaFile)},
		"bucket":           {s3Bucket},
		"X-Amz-Credential": {fmt.Sprintf("%s/%s/ca-central-1/s3/aws4_request", accessKeyID, time.Now().Format("20060102"))},
		"X-Amz-Expires":    {"3600"},
		"x-id":             {"PutObject"},
		"x-amz-acl":        {"public-read-write"},
		"X-Amz-Algorithm":  {"AWS4-HMAC-SHA256"},
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.URL.RawQuery = params.Encode()
	putURL, _, err := httpSigner.PresignHTTP(ftCtx.Context, creds, req, "UNSIGNED-PAYLOAD", "s3", "ca-central-1", time.Now())
	if err != nil {
		return "", "", err
	}
	return putURL, getURL, nil
}

func main() {
	fmt.Print("Starting FolkCreate")
	lambda.Start(Handler)
	fmt.Print("Started FolkCreate")
}
