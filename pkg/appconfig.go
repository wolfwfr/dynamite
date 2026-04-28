package appconfig

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Config struct {
	Profile          *string
	URL              *string
	Region           string
	AvailableRegions []string
	StarredRegions   []string
	Client           *dynamodb.Client
	MaxTables        int
}
