package main

import (
	"context"
	"encoding/base64"
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
	"github.com/sowens-csd/folktells-server/ftdb"
)

const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type mediaAccessResponse struct {
	PutURL    string `json:"putURL",omitempty`
	GetURL    string `json:"getURL",omitempty`
	ExpiresAt int    `json:"expiresAt"`
}

type mediaFileReference struct {
	MediaFile   string `json:"mediaFile", dynamodbav:"mediaFile"`
	ContentType string `json:"contentType", dynamodbav:"contentType"`
	CreatedAt   int    `json:"createdAt", dynamodbav:"createdAt"`
	CreatedBy   string `json:"createdBy", dynamodbav:"createdBy"`
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

	mediaCategory := request.PathParameters["mediaCategory"]
	if mediaCategory != "user" {
		ftCtx.RequestLogger.Info().Str("mediaCategory", mediaCategory).Msg("Unrecognized media category")
		return awsproxy.HandleError(fmt.Errorf("Unrecognized media category %s", mediaCategory), ftCtx.RequestLogger), nil
	}

	mediaReferenceBytes, err := base64.URLEncoding.DecodeString(request.PathParameters["mediaReference"])
	if err != nil {
		ftCtx.RequestLogger.Info().Err(err).Msg("Failed to decode media file")
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	mediaReference := string(mediaReferenceBytes)
	// contentType is only provided in the POST case so it is used to differentiate
	// the two cases since for some reason AWS doesn't provide the http method in the request
	contentType := request.PathParameters["contentType"]
	if len(contentType) == 0 {
		return getMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	} else {
		return createMediaAccessURL(ftCtx, s3Bucket, mediaCategory, mediaReference, request)
	}
}

func getMediaAccessURL(ftCtx awsproxy.FTContext, s3Bucket, mediaCategory, mediaReference string, request awsproxy.Request) (awsproxy.Response, error) {
	resourceID := ftdb.ResourceIDFromUserID((mediaReference))
	referenceID := ftdb.ReferenceIDFromMediaReference(mediaReference)
	var mediaFile mediaFileReference
	ok, err := ftdb.GetItem(ftCtx, resourceID, referenceID, &mediaFile)
	if err != nil {
		ftCtx.RequestLogger.Info().Err(err).Msg("Failed to get media file")
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	if ok {
		expireSeconds := 2 * 60 * 60
		getURL, err := presignGet(ftCtx, s3Bucket, mediaFile.MediaFile, expireSeconds)
		if nil != err {
			ftCtx.RequestLogger.Info().Err(err).Msg("Failed to presign get")
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		return awsproxy.NewJSONResponse(ftCtx, mediaAccessResponse{GetURL: getURL, ExpiresAt: int(time.Now().Add(time.Duration(expireSeconds) * time.Second).UnixMilli())}), nil
	} else {
		return awsproxy.HandleError(fmt.Errorf("No media file found"), ftCtx.RequestLogger), nil
	}
}

func createMediaAccessURL(ftCtx awsproxy.FTContext, s3Bucket, mediaCategory, mediaReference string, request awsproxy.Request) (awsproxy.Response, error) {
	contentTypeBytes, err := base64.URLEncoding.DecodeString(request.PathParameters["contentType"])
	if err != nil {
		ftCtx.RequestLogger.Info().Err(err).Msg("Failed to decode content type")
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	contentType := string(contentTypeBytes)
	resourceID := ftdb.ResourceIDFromUserID((mediaReference))
	referenceID := ftdb.ReferenceIDFromMediaReference(mediaReference)
	var mediaFile mediaFileReference
	ok, err := ftdb.GetItem(ftCtx, resourceID, referenceID, &mediaFile)
	if err != nil {
		ftCtx.RequestLogger.Info().Err(err).Msg("Failed to get media file")
		return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
	}
	expireSeconds := 30 * 60
	if ok {
		putURL, err := presignPut(ftCtx, s3Bucket, mediaFile.MediaFile, contentType, expireSeconds)
		if nil != err {
			ftCtx.RequestLogger.Info().Err(err).Msg("Failed to presign get")
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		return awsproxy.NewJSONResponse(ftCtx, mediaAccessResponse{PutURL: putURL, ExpiresAt: int(time.Now().Add(time.Duration(expireSeconds) * time.Second).UnixMilli())}), nil
	} else {
		mediaFilename := fmt.Sprintf("%s_%s.%s", ftdb.NewUUID(), "profile", "jpg")
		mediaFile = mediaFileReference{MediaFile: mediaFilename, CreatedAt: ftdb.NowMillisecondsSinceEpoch(), CreatedBy: ftCtx.UserID, ContentType: contentType}
		err := ftdb.PutItem(ftCtx, resourceID, referenceID, &mediaFile)
		if nil != err {
			ftCtx.RequestLogger.Info().Err(err).Msg("Failed to put media file")
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		putURL, err := presignPut(ftCtx, s3Bucket, mediaFile.MediaFile, contentType, expireSeconds)
		if nil != err {
			ftCtx.RequestLogger.Info().Err(err).Msg("Failed to presign put")
			return awsproxy.HandleError(err, ftCtx.RequestLogger), nil
		}
		return awsproxy.NewJSONResponse(ftCtx, mediaAccessResponse{PutURL: putURL, ExpiresAt: int(time.Now().Add(time.Duration(expireSeconds) * time.Second).UnixMilli())}), nil
	}

}

func presignGet(ftCtx awsproxy.FTContext, s3Bucket, mediaFile string, expiresSeconds int) (string, error) {
	accessToken, _ := awsproxy.SharedCredentialParameters(ftCtx.Context)
	uri := fmt.Sprintf("https://sharedstack-folktellsmediabucket60d66dfa-umnguk71tkci.s3.ca-central-1.amazonaws.com/%s", url.PathEscape(mediaFile))
	req, _ := http.NewRequest("GET", uri, nil)
	params := url.Values{
		"X-Amz-Credential": {fmt.Sprintf("%s/%s/ca-central-1/s3/aws4_request", accessToken, time.Now().Format("20060102"))},
		"X-Amz-Expires":    {fmt.Sprintf("%d", expiresSeconds)},
		"x-id":             {"GetObject"},
		"x-amz-acl":        {"public-read"},
		"X-Amz-Algorithm":  {"AWS4-HMAC-SHA256"},
	}

	return presignRequest(ftCtx, req, params)
}

func presignPut(ftCtx awsproxy.FTContext, s3Bucket, mediaFile, contentType string, expiresSeconds int) (string, error) {
	accessToken, _ := awsproxy.SharedCredentialParameters(ftCtx.Context)

	uri := fmt.Sprintf("https://sharedstack-folktellsmediabucket60d66dfa-umnguk71tkci.s3.ca-central-1.amazonaws.com/%s", url.PathEscape(mediaFile))
	req, _ := http.NewRequest("PUT", uri, nil)
	params := url.Values{
		"key":              {fmt.Sprintf("/%s", mediaFile)},
		"bucket":           {s3Bucket},
		"X-Amz-Credential": {fmt.Sprintf("%s/%s/ca-central-1/s3/aws4_request", accessToken, time.Now().Format("20060102"))},
		"X-Amz-Expires":    {fmt.Sprintf("%d", expiresSeconds)},
		"x-id":             {"PutObject"},
		"x-amz-acl":        {"public-read-write"},
		"X-Amz-Algorithm":  {"AWS4-HMAC-SHA256"},
	}
	req.Header.Set("Content-Type", contentType)
	return presignRequest(ftCtx, req, params)
}

func presignRequest(ftCtx awsproxy.FTContext, req *http.Request, params url.Values) (string, error) {
	accessToken, secretKey := awsproxy.SharedCredentialParameters(ftCtx.Context)
	// os.Setenv("AWS_ACCESS_KEY_ID", accessToken)
	// os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	// os.Setenv("AWS_REGION", "ca-central-1")
	// os.Setenv("AWS_DEFAULT_REGION", "ca-central-1")
	ftCtx.RequestLogger.Debug().Int("AccessKey", len(accessToken)).Int("SecretKey", len(secretKey)).Msg("Have params for presign")

	httpSigner := signer.NewSigner()
	appCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessToken, secretKey, ""))
	creds, err := appCreds.Retrieve(ftCtx.Context)
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = params.Encode()
	presigned, _, err := httpSigner.PresignHTTP(ftCtx.Context, creds, req, "UNSIGNED-PAYLOAD", "s3", "ca-central-1", time.Now())
	if err != nil {
		return "", err
	}
	return presigned, nil
}

func main() {
	fmt.Print("Starting FolkCreate")
	lambda.Start(Handler)
	fmt.Print("Started FolkCreate")
}

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
