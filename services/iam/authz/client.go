package iam

import "github.com/marmotedu/medu-sdk-go/sdk"

const (
	defaultEndpoint = "127.0.0.1:9090"
	serviceName     = "iam.authz"
)

type Client struct {
	sdk.Client
}

func NewClient(config *sdk.Config, credential *sdk.Credential) (client *Client, err error) {
	client = &Client{}
	if config == nil {
		config = sdk.NewConfig().WithEndpoint(defaultEndpoint)
	}

	client.Init(serviceName).WithCredential(credential).WithConfig(config)
	return
}

func NewClientWithSecret(secretID, secretKey string) (client *Client, err error) {
	client = &Client{}
	config := sdk.NewConfig().WithEndpoint(defaultEndpoint)
	client.Init(serviceName).WithSecret(secretID, secretKey).WithConfig(config)
	return
}
