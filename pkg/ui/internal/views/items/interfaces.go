package itemselection

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/gen.go -package=mocks
type dynamodbClient interface {
	ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error)
	QueryTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error)
}
