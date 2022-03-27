package notification

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/sowens-csd/ftlambdas/awsproxy"
)

var snsClient *sns.Client

func getSNSClient(ftCtx awsproxy.FTContext) *sns.Client {
	if nil == snsClient {
		cfg, err := config.LoadDefaultConfig(ftCtx.Context, config.WithRegion("us-east-1"))
		if nil != err {
			panic("Could not load config")
		}
		snsClient = sns.NewFromConfig(cfg)
	}
	return snsClient
}
