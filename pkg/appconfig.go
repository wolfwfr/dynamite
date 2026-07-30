package appconfig

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type Config struct {
	Profile          *string
	URL              *string
	Region           string
	AvailableRegions []string
	StarredRegions   []string
	Client           *dynamodb.Client

	// view-specific settings
	Tables Tables
	Items  Items

	// credentials
	MFACredentialCB func() (string, error)
	MFACredentialC  chan<- CredentialsResponse
}

type Tables struct {
	MaxTables    int
	Pagesize     int
	PrimaryWidth int
}

type Items struct {
	PrimaryWidth int
	PageSize     int
}

type CredentialsRequest struct{}

type CredentialsResponse struct {
	Token string
	Error error
}
