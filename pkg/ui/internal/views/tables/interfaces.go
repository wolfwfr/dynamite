package tableselection

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

type dynamodbClient interface {
	ListTables(client *dynamodb.Client, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error)
	DescribeTable(client *dynamodb.Client, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error)
}
