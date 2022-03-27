package awsproxy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// ParameterStore simplest behaviour to get and set string parameters
type ParameterStore interface {
	GetParameter(ctx context.Context, name string) (string, error)
	PutParameter(ctx context.Context, name string, value string) error
}

type awsParameterStore struct {
	ParameterStore
	Service *ssm.Client
}

// NewParameterStore returns an instance of ParameterStore
func NewParameterStore(cfg aws.Config) ParameterStore {
	paramstore := ssm.NewFromConfig(cfg)
	return awsParameterStore{Service: paramstore}
}

// GetParameter return the value of a parameter from the store
func (ps awsParameterStore) GetParameter(ctx context.Context, name string) (string, error) {
	withDecryption := true
	param, err := ps.Service.GetParameter(ctx, &ssm.GetParameterInput{Name: &name, WithDecryption: withDecryption})
	if nil != err {
		return "", err
	}
	return *param.Parameter.Value, nil
}

// PutParameter update the value of a parameter in the store
func (ps awsParameterStore) PutParameter(ctx context.Context, name string, value string) error {
	overwrite := true
	_, err := ps.Service.PutParameter(ctx, &ssm.PutParameterInput{Name: &name, Overwrite: overwrite, Type: types.ParameterTypeSecureString, Value: &value})
	return err
}
