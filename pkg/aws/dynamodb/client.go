package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewClient(cfg *aws.Config, url *string) *dynamodb.Client {
	opts := []func(*dynamodb.Options){}
	if url != nil && *url != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(*url)
		})
	}
	client := dynamodb.NewFromConfig(*cfg, opts...)
	return client
}
