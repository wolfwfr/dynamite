package tableselection

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/gen.go -package=mocks
type dynamodbClient interface {
	ListTables(client *dynamodb.Client, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error)
	DescribeTable(client *dynamodb.Client, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error)
}
