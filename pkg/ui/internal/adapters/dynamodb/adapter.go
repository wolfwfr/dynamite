// Adapter parses dynamodb connector responses for UI display purposes,
// including JSON/YAML and styling
package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/parsing"
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

// ScanTable forwards the call to the dynamodb connector & parses the results
// for UI display.
func ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error) {
	dparams := cncrtypes.ScanParameters{
		KeyDetails:       params.KeyDetails,
		IndexName:        params.IndexName,
		KeySchema:        params.KeySchema,
		Limit:            params.Limit,
		LastEvaluatedKey: params.LastEvaluatedKey,
	}
	out, err := connector.ScanTable(client, ctx, table, dparams)
	if out == nil || err != nil {
		return nil, err
	}

	res := &apitypes.ScanResponse{
		Items: apitypes.Items{
			JSON:       make([]string, 0, len(out.Items)),
			JSONStyled: make([]string, 0, len(out.Items)),
			YAML:       make([]string, 0, len(out.Items)),
			YAMLStyled: make([]string, 0, len(out.Items)),
			Raw:        out.Items,
			TableKeys:  make([][]apitypes.KeyValue, 0, len(out.Items)),
		},
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)

	// TODO: reconsider parsing to both JSON & YAML all the time
	for _, item := range out.Items {
		yaml := parsing.ParseItemToYAML(item, *hkey, rkey)
		json, jsonStyled, keys := parsing.NewJSONParser().ParseToJSONWithKeys(item, *hkey, rkey)
		res.Items.JSON = append(res.Items.JSON, json)
		res.Items.JSONStyled = append(res.Items.JSONStyled, jsonStyled)
		res.Items.YAML = append(res.Items.YAML, yaml)
		res.Items.YAMLStyled = append(res.Items.YAMLStyled, yaml)
		res.Items.TableKeys = append(res.Items.TableKeys, keys)
	}

	return res, err
}

// QueryTable forwards the call to the dynamodb connector & parses the results
// for UI display.
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
	if out == nil || err != nil {
		return nil, err
	}

	res := &apitypes.QueryResponse{
		Items: apitypes.Items{
			JSON:      make([]string, 0, len(out.Items)),
			YAML:      make([]string, 0, len(out.Items)),
			Raw:       out.Items,
			TableKeys: make([][]apitypes.KeyValue, 0, len(out.Items)),
		},
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)

	// TODO: reconsider parsing to both JSON & YAML all the time
	for _, item := range out.Items {
		yaml := parsing.ParseItemToYAML(item, *hkey, rkey)
		json, _, keys := parsing.NewJSONParser().ParseToJSONWithKeys(item, *hkey, rkey)
		res.Items.JSON = append(res.Items.JSON, json)
		res.Items.YAML = append(res.Items.YAML, yaml)
		res.Items.TableKeys = append(res.Items.TableKeys, keys)
	}
	return res, nil
}

func parsePrimaryKeys(schema []types.KeySchemaElement) (*string, *string) {
	var hash *string
	var rang *string

	// obtain key names
	for _, k := range schema {
		if k.KeyType == types.KeyTypeHash {
			hash = k.AttributeName
		} else if k.KeyType == types.KeyTypeRange {
			rang = k.AttributeName
		}
	}

	return hash, rang
}
