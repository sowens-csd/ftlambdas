module github.com/sowens-csd/folktells-cloud-deploy/lambdas/signup_verify

go 1.16

require (
	github.com/aws/aws-lambda-go v1.32.0
	github.com/aws/aws-sdk-go-v2/config v1.15.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi v1.10.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ses v1.14.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.17.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssm v1.27.2 // indirect
	github.com/rs/zerolog v1.27.0 // indirect
	github.com/sowens-csd/folktells-server v1.1.11
	github.com/sowens-csd/ftlambdas v0.0.0-20220313160211-a7b1abb84b84
	golang.org/x/sys v0.0.0-20220615213510-4f61da869c0c // indirect
)
