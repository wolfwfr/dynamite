package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

// TODO: add filters everywhere

type dynamodbClient interface {
	ListTables(context.Context, *dynamodb.ListTablesInput, ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	DescribeTable(context.Context, *dynamodb.DescribeTableInput, ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	Scan(context.Context, *dynamodb.ScanInput, ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// ListTables lists the tables available to the specified dynamodb-client. It
// returns a response containing table names and optionally a pagination-key.
// Note that only up to 100 tables can be retrieved at once.
func ListTables(client dynamodbClient, ctx context.Context, req apitypes.ListTablesRequest) (*apitypes.ListTablesResponse, error) {
	p := dynamodb.ListTablesInput{
		ExclusiveStartTableName: req.LastEvaluatedTableName,
		Limit:                   req.Limit,
	}
	out, err := client.ListTables(ctx, &p)
	if err != nil {
		return nil, err
	}
	resp := &apitypes.ListTablesResponse{
		TableNames:             out.TableNames,
		LastEvaluatedTableName: out.LastEvaluatedTableName,
	}
	return resp, nil
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
	var index *string
	if params.IndexName != nil && *params.IndexName != "" {
		index = params.IndexName
	}
	p := dynamodb.ScanInput{
		TableName:         &table,
		Limit:             toPtr(int32(params.Limit)),
		ExclusiveStartKey: params.LastEvaluatedKey,
		IndexName:         index,

		// ScanFilter:                map[string]types.Condition{},
		// AttributesToGet:           []string{},
		// ConditionalOperator:       "",
		// ConsistentRead:            new(bool),
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
		Items:            out.Items,
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	return res, nil
}

func QueryTable(client dynamodbClient, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error) {
	keys, values, names, err := formatQueryKeys(params)
	if err != nil {
		return nil, err
	}
	var index *string
	if params.IndexName != nil && *params.IndexName != "" {
		index = params.IndexName
	}
	ascendingOrder := !params.Descending
	p := dynamodb.QueryInput{
		TableName:                 &table,
		Limit:                     toPtr(int32(params.Limit)),
		KeyConditionExpression:    &keys,
		ExpressionAttributeValues: values,
		ExpressionAttributeNames:  names,
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
		Items:            out.Items,
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	return res, nil
}

func toPtr[T any](in T) *T {
	return &in
}
