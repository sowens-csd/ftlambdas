module github.com/sowens-csd/ftlambdas/ftlambdas

go 1.16

require (
	github.com/ReneKroon/ttlcache v1.7.0
	github.com/aws/aws-lambda-go v1.32.0
	github.com/aws/aws-sdk-go-v2 v1.16.1
	github.com/aws/aws-sdk-go-v2/config v1.15.0
	github.com/aws/aws-sdk-go-v2/credentials v1.10.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.8.0
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.10.2
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.15.0
	github.com/aws/aws-sdk-go-v2/service/firehose v1.14.0
	github.com/aws/aws-sdk-go-v2/service/kinesisvideo v1.4.1
	github.com/aws/aws-sdk-go-v2/service/kinesisvideosignaling v1.4.1
	github.com/aws/aws-sdk-go-v2/service/ses v1.14.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.17.1
	github.com/aws/aws-sdk-go-v2/service/ssm v1.22.0
	github.com/aws/smithy-go v1.11.2
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/plivo/plivo-go v7.2.0+incompatible
	github.com/rs/zerolog v1.26.1
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064
	golang.org/x/sys v0.0.0-20220319134239-a9b59b0215f8 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
)
