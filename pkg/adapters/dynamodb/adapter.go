// Adapter parses dynamodb connector responses for UI display purposes,
// including JSON/YAML and styling
package dynamodb

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/lib/styles"
	"github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/parsing"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/logging"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// Adapter encapsulates the dynamo-db adapter functions. Although stateless, it
// can be mocked or decorated.
type Adapter struct {
	logger  *slog.Logger
	styling apitypes.ObjectStyling
}

// NewAdapter returns a new instance of Adapter.
func NewAdapter(logger *slog.Logger, opts ...Option) *Adapter {
	options := &options{}
	for _, o := range opts {
		o(options)
	}
	return &Adapter{
		logger:  logger.With(slog.String(logging.ComponentKey, "dynamodb-adapter")),
		styling: options.objectStyling,
	}
}

// ListTables lists the tables available to the specified dynamodb-client. It
// returns a response containing table names and optionally a pagination-key.
// Note that only up to 100 tables can be retrieved at once.
func (a *Adapter) ListTables(client *dynamodb.Client, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error) {
	dreq := dynamodb.ListTablesInput{
		ExclusiveStartTableName: req.LastEvaluatedTableName,
		Limit:                   req.Limit,
	}
	res, err := client.ListTables(ctx, &dreq)
	if res == nil {
		return nil, err
	}
	return &apitypes.ListTablesResponse{
		TableNames:             res.TableNames,
		LastEvaluatedTableName: res.LastEvaluatedTableName,
	}, err
}

// DescribeTable describes the specified table and returns a curated
// table-details response. If the table could not be found it wraps the original
// aws error.
func (a *Adapter) DescribeTable(client *dynamodb.Client, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error) {
	p := dynamodb.DescribeTableInput{
		TableName: &tableName,
	}
	out, err := client.DescribeTable(ctx, &p)
	if err != nil {
		return nil, fmt.Errorf("describing table: %w", err)
	}
	return &apitypes.DescribeTableResponse{
		AttributeDefinitions:      out.Table.AttributeDefinitions,
		BillingModeSummary:        out.Table.BillingModeSummary,
		CreationDateTime:          out.Table.CreationDateTime,
		DeletionProtectionEnabled: out.Table.DeletionProtectionEnabled,
		GlobalSecondaryIndexes:    out.Table.GlobalSecondaryIndexes,
		ItemCount:                 out.Table.ItemCount,
		KeySchema:                 out.Table.KeySchema,
		LocalSecondaryIndexes:     out.Table.LocalSecondaryIndexes,
		OnDemandThroughput:        out.Table.OnDemandThroughput,
		ProvisionedThroughput:     out.Table.ProvisionedThroughput,
		SSEDescription:            out.Table.SSEDescription,
		TableArn:                  out.Table.TableArn,
		TableClassSummary:         out.Table.TableClassSummary,
		TableId:                   out.Table.TableId,
		TableName:                 out.Table.TableName,
		TableSizeBytes:            out.Table.TableSizeBytes,
	}, err
}

// ScanTable scans the next page from the specified table, parses the results
// for UI display.
func (a *Adapter) ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error) {
	var index *string
	if params.IndexName != nil && *params.IndexName != "" {
		index = params.IndexName
	}

	fex, fnm, fvl := a.buildFilterExpression(params.FilterParameters)

	p := dynamodb.ScanInput{
		TableName:                 &table,
		Limit:                     u.ToPtr(int32(params.Limit)),
		ExclusiveStartKey:         params.LastEvaluatedKey,
		IndexName:                 index,
		FilterExpression:          fex,
		ExpressionAttributeNames:  fnm,
		ExpressionAttributeValues: fvl,

		// AttributesToGet:           []string{},
		// ConditionalOperator:       "",
		// ConsistentRead:            new(bool),
		// ProjectionExpression:      new(string),
		// ReturnConsumedCapacity:    "",
		// Segment:                   new(int32),
		// Select:                    "",
		// TotalSegments:             new(int32),
	}

	out, err := client.Scan(ctx, &p)
	if err != nil {
		return nil, fmt.Errorf("scanning table: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("expected output, but dynamo-db query returned 'nil'")
	}

	res := &apitypes.ScanResponse{
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)
	res.Items = a.parseItems(out.Items, hkey, rkey)

	return res, err
}

// QueryTable obtains the next page from the specified table, parses the results
// for UI display.
func (a *Adapter) QueryTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error) {
	keys, values, names, err := formatQueryKeys(params)
	if err != nil {
		return nil, err
	}
	var index *string
	if params.IndexName != nil && *params.IndexName != "" {
		index = params.IndexName
	}

	fex, fnm, fvl := a.buildFilterExpression(params.FilterParameters)

	var rem1 map[string]string
	names, rem1 = u.MergeMapsSafe(names, fnm)
	if len(rem1) > 0 {
		panic("BUG: expression_attribute_names name-collision; Please report to maintainer")
	}
	var rem2 map[string]types.AttributeValue
	values, rem2 = u.MergeMapsSafe(values, fvl)
	if len(rem2) > 0 {
		panic("BUG: expression_attribute_values name-collision; Please report to maintainer")
	}

	ascendingOrder := !params.Descending
	p := dynamodb.QueryInput{
		TableName:                 &table,
		Limit:                     u.ToPtr(int32(params.Limit)),
		KeyConditionExpression:    &keys,
		FilterExpression:          fex,
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
		Select:                    "ALL_ATTRIBUTES",
		IndexName:                 index,
		ExclusiveStartKey:         params.LastEvaluatedKey,
		ScanIndexForward:          &ascendingOrder,

		// AttributesToGet:           []string{},
		// ConsistentRead:            new(bool),
		// FilterExpression:          new(string),
		// ProjectionExpression:      new(string),
		// QueryFilter:               map[string]types.Condition{},
		// ReturnConsumedCapacity:    "",
	}

	out, err := client.Query(ctx, &p)
	if err != nil {
		return nil, fmt.Errorf("querying table: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("expected output, but dynamo-db query returned 'nil'")
	}

	res := &apitypes.QueryResponse{
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	hkey, rkey := parsePrimaryKeys(params.KeySchema)
	res.Items = a.parseItems(out.Items, hkey, rkey)

	return res, nil
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
