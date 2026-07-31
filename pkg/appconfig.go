package appconfig

import (
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Config struct {
	Logger *slog.Logger

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

	// initialisation (CLI flags)
	Initialisation Initialisation
}

type Initialisation struct {
	Table       string
	Index       string
	Query       Queryinitialisation
	ViewOptions ViewOptionsInitialisation
}

type Queryinitialisation struct {
	HashkeyValue     string
	RangekeyValue1   *string
	RangekeyValue2   *string
	RangeKeyOperator string
	RangeDescending  bool
}

// TODO: consider
type ViewOptionsInitialisation struct {
}

type Tables struct {
	MaxTables       int
	Pagesize        int
	PrimaryWidth    int
	HighlightRegexp []string
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
