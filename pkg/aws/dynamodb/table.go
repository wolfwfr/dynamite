package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/wolfwfr/dynamite/pkg/aws/dynamodb/parsing"
	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

// TODO: add filters everywhere
// TODO: add pagination everywhere (incl. pagesize, pagekey)

type dynamodbClient interface {
	ListTables(context.Context, *dynamodb.ListTablesInput, ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	DescribeTable(context.Context, *dynamodb.DescribeTableInput, ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	Scan(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// ListTables lists the tables available to the specified dynamodb-client. It
// simply returns a `[]string` with the table-names.
func ListTables(client dynamodbClient, ctx context.Context) ([]string, error) {
	p := dynamodb.ListTablesInput{
		// ExclusiveStartTableName: new(string),
		// Limit:                   new(int32),
	}
	out, err := client.ListTables(ctx, &p)
	if err != nil {
		return nil, err
	}
	return out.TableNames, nil
}

// DescribeTable describes the specified table and returns a curated
// table-details response. If the table could not be found it wraps the original
// aws error.
func DescribeTable(client dynamodbClient, ctx context.Context, tableName string) (*apitypes.DescribeTableResponse, error) {
	p := dynamodb.DescribeTableInput{
		TableName: &tableName,
	}

	out, err := client.DescribeTable(ctx, &p)
	if err != nil {
		return nil, fmt.Errorf("describing table: %w", err)
	}
	res := &apitypes.DescribeTableResponse{
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
	}
	return res, nil
}

func ScanTable(client dynamodbClient, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error) {
	hkey, rkey := parsePrimaryKeys(params.KeySchema) // TODO: prevent waste
	p := dynamodb.ScanInput{
		TableName: &table,
		Limit:     toPtr(int32(params.Limit)),
		// Limit:                     new(int32),
		// ScanFilter:                map[string]types.Condition{},
		// IndexName:                 new(string),

		// AttributesToGet:           []string{},
		// ConditionalOperator:       "",
		// ConsistentRead:            new(bool),
		// ExclusiveStartKey:         map[string]types.AttributeValue{},
		// ExpressionAttributeNames:  map[string]string{},
		// ExpressionAttributeValues: map[string]types.AttributeValue{},
		// FilterExpression:          new(string),
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
		Items: apitypes.Items{
			JSON:      make([]string, 0, len(out.Items)),
			YAML:      make([]string, 0, len(out.Items)),
			Raw:       out.Items,
			TableKeys: make([][]apitypes.KeyValue, 0, len(out.Items)),
		},
	}

	// TODO: reconsider parsing to both JSON & YAML all the time
	for _, item := range out.Items {
		yaml := parsing.ParseItemToYAML(item, *hkey, rkey)
		json, keys := parsing.ParseToJSONWithKeys(item, *hkey, rkey)
		res.Items.JSON = append(res.Items.JSON, json)
		res.Items.YAML = append(res.Items.YAML, yaml)
		res.Items.TableKeys = append(res.Items.TableKeys, keys)
	}

	return res, nil
}

// TODO: add options for rangekey comparisons (e.g. 'eq', 'gt', 'between', etc.)
func QueryTable(client dynamodbClient, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error) {
	hkey, rkey, keys, values := formatQueryKeys(params)
	p := dynamodb.QueryInput{
		TableName:                 &table,
		KeyConditionExpression:    &keys,
		ExpressionAttributeValues: values,
		Select:                    "ALL_ATTRIBUTES",

		// AttributesToGet:           []string{},
		// ConsistentRead:            new(bool),
		// ExclusiveStartKey:         map[string]types.AttributeValue{},
		// ExpressionAttributeNames:  map[string]string{},
		// ExpressionAttributeValues: map[string]types.AttributeValue{},
		// FilterExpression:          new(string),
		// IndexName:                 new(string),
		// Limit:                     new(int32),
		// ProjectionExpression:      new(string),
		// QueryFilter:               map[string]types.Condition{},
		// ReturnConsumedCapacity:    "",
		// ScanIndexForward:          new(bool),
	}

	out, err := client.Query(ctx, &p)
	if err != nil {
		return nil, fmt.Errorf("querying table: %w", err)
	}
	if out == nil {
		return nil, fmt.Errorf("expected output, but dynamo-db query returned 'nil'")
	}
	res := &apitypes.QueryResponse{
		Items: apitypes.Items{
			JSON:      make([]string, 0, len(out.Items)),
			YAML:      make([]string, 0, len(out.Items)),
			Raw:       out.Items,
			TableKeys: make([][]apitypes.KeyValue, 0, len(out.Items)),
		},
	}

	// TODO: reconsider parsing to both JSON & YAML all the time
	for _, item := range out.Items {
		yaml := parsing.ParseItemToYAML(item, hkey, rkey)
		json, keys := parsing.ParseToJSONWithKeys(item, hkey, rkey)
		res.Items.JSON = append(res.Items.JSON, json)
		res.Items.YAML = append(res.Items.YAML, yaml)
		res.Items.TableKeys = append(res.Items.TableKeys, keys)
	}

	return res, nil
}

func toPtr[T any](in T) *T {
	return &in
}
