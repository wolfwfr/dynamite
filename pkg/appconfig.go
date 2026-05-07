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

	// credentials
	MFACredentialCB func() (string, error)
	MFACredentialC  chan<- CredentialsResponse
}

type CredentialsRequest struct{}

type CredentialsResponse struct {
	Token string
	Error error
}
