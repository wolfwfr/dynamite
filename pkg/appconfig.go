package appconfig

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Config struct {
	Profile *string
	Region  string
	Client  *dynamodb.Client
}
