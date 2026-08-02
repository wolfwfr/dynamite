// Adapter parses dynamodb connector responses for UI display purposes,
// including JSON/YAML and styling
package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/lib/styles"
	"github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/parsing"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"

	connector "github.com/wolfwfr/dynamite/pkg/aws/dynamodb"
	cncrtypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

// Adapter encapsulates the dynamo-db adapter functions. Although stateless, it
// can be mocked or decorated.
type Adapter struct {
	styling apitypes.ObjectStyling
}

// NewAdapter returns a new instance of Adapter.
func NewAdapter(opts ...Option) *Adapter {
	options := &options{}
	for _, o := range opts {
		o(options)
	}
	return &Adapter{
		styling: options.objectStyling,
	}
}

// simple one on one translation
func (a *Adapter) ListTables(client *dynamodb.Client, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error) {
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
func (a *Adapter) DescribeTable(client *dynamodb.Client, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error) {
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
func (a *Adapter) ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error) {
	dparams := cncrtypes.ScanParameters{
		KeyDetails:       params.KeyDetails,
		IndexName:        params.IndexName,
		KeySchema:        params.KeySchema,
		FilterParameters: convertFilterParameters(params.FilterParameters),
		Limit:            params.Limit,
		LastEvaluatedKey: params.LastEvaluatedKey,
	}
	out, err := connector.ScanTable(client, ctx, table, dparams)
	if out == nil || err != nil {
		return nil, err
	}

	res := &apitypes.ScanResponse{
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)
	res.Items = a.parseItems(out.Items, hkey, rkey)

	return res, err
}

// QueryTable forwards the call to the dynamodb connector & parses the results
// for UI display.
func (a *Adapter) QueryTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error) {
	dparams := cncrtypes.QueryParameters{
		KeyDetails:       params.KeyDetails,
		IndexName:        params.IndexName,
		KeySchema:        params.KeySchema,
		FilterParameters: convertFilterParameters(params.FilterParameters),
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
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)
	res.Items = a.parseItems(out.Items, hkey, rkey)

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

func (a *Adapter) parseItems(raw []map[string]types.AttributeValue, hkey, rkey *string) apitypes.Items {
	items := apitypes.Items{
		JSON:       make([]string, 0, len(raw)),
		JSONStyled: make([]styles.ObjectStyle, 0, len(raw)),
		YAML:       make([]string, 0, len(raw)),
		YAMLStyled: make([]styles.ObjectStyle, 0, len(raw)),
		Raw:        raw,
		TableKeys:  make([][]apitypes.KeyValue, 0, len(raw)),
	}

	// TODO: reconsider parsing to both JSON & YAML all the time
	for _, item := range raw {
		yaml, yamlStyled := parsing.NewYAMLParser(a.styling).ParseItemToYAML(item, *hkey, rkey)
		json, jsonStyled, keys := parsing.NewJSONParser(a.styling).ParseToJSONWithKeys(item, *hkey, rkey)
		items.JSON = append(items.JSON, json)
		items.JSONStyled = append(items.JSONStyled, jsonStyled)
		items.YAML = append(items.YAML, yaml)
		items.YAMLStyled = append(items.YAMLStyled, yamlStyled)
		items.TableKeys = append(items.TableKeys, keys)
	}
	return items
}
