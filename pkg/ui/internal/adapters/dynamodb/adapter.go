// Adapter parses dynamodb connector responses for UI display purposes,
// including JSON/YAML and styling
package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"

	connector "github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	cncrtypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

// simple one on one translation
func ListTables(client *dynamodb.Client, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error) {
	dreq := cncrtypes.ListTablesRequest{
		LastEvaluatedTableName: req.LastEvaluatedTableName,
		Limit:                  req.Limit,
	}
	res, err := connector.ListTables(client, ctx, dreq)
	if res == nil {
		return nil, err
	}
	return &apitypes.ListTablesResponse{
		TableNames:             res.TableNames,
		LastEvaluatedTableName: res.LastEvaluatedTableName,
	}, err
}

// simple one on one translation
func DescribeTable(client *dynamodb.Client, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error) {
	res, err := connector.DescribeTable(client, ctx, tableName)
	if res == nil {
		return nil, err
	}
	return &apitypes.DescribeTableResponse{
		AttributeDefinitions:      res.AttributeDefinitions,
		BillingModeSummary:        res.BillingModeSummary,
		CreationDateTime:          res.CreationDateTime,
		DeletionProtectionEnabled: res.DeletionProtectionEnabled,
		GlobalSecondaryIndexes:    res.GlobalSecondaryIndexes,
		ItemCount:                 res.ItemCount,
		KeySchema:                 res.KeySchema,
		LocalSecondaryIndexes:     res.LocalSecondaryIndexes,
		OnDemandThroughput:        res.OnDemandThroughput,
		ProvisionedThroughput:     res.ProvisionedThroughput,
		SSEDescription:            res.SSEDescription,
		TableArn:                  res.TableArn,
		TableClassSummary:         res.TableClassSummary,
		TableId:                   res.TableId,
		TableName:                 res.TableName,
		TableSizeBytes:            res.TableSizeBytes,
	}, err
}

func ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error) {
	dparams := cncrtypes.ScanParameters{
		KeyDetails:       params.KeyDetails,
		IndexName:        params.IndexName,
		KeySchema:        params.KeySchema,
		Limit:            params.Limit,
		LastEvaluatedKey: params.LastEvaluatedKey,
	}
	res, err := connector.ScanTable(client, ctx, table, dparams)
	if res == nil {
		return nil, err
	}

	tableKeys := make([][]apitypes.KeyValue, len(res.Items.TableKeys))
	for i, kk := range res.Items.TableKeys {
		v := make([]apitypes.KeyValue, len(kk))
		for j, k := range kk {
			v[j] = apitypes.KeyValue{
				Key:   k.Key,
				Value: k.Value,
			}
		}
		tableKeys[i] = v
	}

	items := apitypes.Items{
		JSON:      res.Items.JSON,
		YAML:      res.Items.YAML,
		Raw:       res.Items.Raw,
		TableKeys: tableKeys,
	}

	return &apitypes.ScanResponse{
		Items:            items,
		LastEvaluatedKey: res.LastEvaluatedKey,
	}, err
}

func QueryTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error) {
	dparams := cncrtypes.QueryParameters{
		KeyDetails:       params.KeyDetails,
		IndexName:        params.IndexName,
		KeySchema:        params.KeySchema,
		Limit:            params.Limit,
		LastEvaluatedKey: params.LastEvaluatedKey,
		HashKeyValue:     params.HashKeyValue,
		RangeKeyValue1:   params.RangeKeyValue1,
		RangeKeyValue2:   params.RangeKeyValue2,
		RangeKeyOperator: cncrtypes.RangeKeyOperator(params.RangeKeyOperator),
		Descending:       params.Descending,
	}
	out, err := connector.QueryTable(client, ctx, table, dparams)
	if out == nil {
		return nil, err
	}

	tableKeys := make([][]types.KeyValue, len(out.Items.TableKeys))
	for i, kk := range out.Items.TableKeys {
		v := make([]types.KeyValue, len(kk))
		for j, k := range kk {
			v[j] = types.KeyValue{
				Key:   k.Key,
				Value: k.Value,
			}
		}
		tableKeys[i] = v
	}

	items := types.Items{
		JSON:      out.Items.JSON,
		YAML:      out.Items.YAML,
		Raw:       out.Items.Raw,
		TableKeys: tableKeys,
	}

	return &types.QueryResponse{
		Items:            items,
		LastEvaluatedKey: out.LastEvaluatedKey,
	}, err
}
