package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/util"
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

	fex, fnm, fvl := buildFilterExpression(params.FilterParameters)

	p := dynamodb.ScanInput{
		TableName:                 &table,
		Limit:                     toPtr(int32(params.Limit)),
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
		Items:            out.Items,
		LastEvaluatedKey: out.LastEvaluatedKey,
	}

	return res, nil
}

func buildFilterExpression(params []apitypes.FilterExpressionParameters) (expr *string, exprNames map[string]string, exprVals map[string]types.AttributeValue) {
	if len(params) == 0 {
		return nil, nil, nil
	}
	exprNames = make(map[string]string)
	exprVals = make(map[string]types.AttributeValue)

	nameAliasConstructor := newExpressionNameAliasConstructor()
	valueAliasConstructor := newExpressionValueAliasConstructor()

	var exprr string
	for i, p := range params {
		var s string
		if p.AttributeName == "" || p.AttributeValue1 == "" || p.Operator == apitypes.Noop_F || p.Operator == apitypes.Between_F && util.IfNotNil(p.AttributeValue2, "") == "" {
			continue
		}

		attrNameAlias := nameAliasConstructor(p.AttributeName)
		attrValueAlias := valueAliasConstructor(p.AttributeValue1, p.AttributeValueType)
		exprNames[attrNameAlias] = p.AttributeName
		exprVals[attrValueAlias] = ToAttrValue(p.AttributeValue1, p.AttributeValueType)

		if i > 0 {
			s = " AND " // TODO: consider supporting 'OR' operator
		}

		s = fmt.Sprintf("%s%s %s %s", s, attrNameAlias, ParseFilterOperator(p.Operator), attrValueAlias)

		if p.Operator == apitypes.Between_F {
			attrVal2 := util.IfNotNil(p.AttributeValue2, "")
			attrValue2Alias := valueAliasConstructor(attrVal2, p.AttributeValueType)
			exprVals[attrValue2Alias] = ToAttrValue(attrVal2, p.AttributeValueType)
			s = fmt.Sprintf("%s AND %s", s, attrValue2Alias)
		}
		exprr = fmt.Sprintf("%s%s", exprr, s)
	}
	expr = &exprr
	return
}

func newExpressionNameAliasConstructor() func(attrName string) string {
	names := make(map[rune]int)
	namesSeen := make(map[string]string)

	return func(attrName string) string {
		if nm, ok := namesSeen[attrName]; ok {
			return nm
		}

		// alias letter
		nl := rune(attrName[0])

		// alias number
		names[nl] = names[nl] + 1

		attrNameAlias := fmt.Sprintf("#%s%d", string(nl), names[nl])

		namesSeen[attrName] = attrNameAlias

		return attrNameAlias
	}
}

func newExpressionValueAliasConstructor() func(attrValue string, attrValueType types.ScalarAttributeType) string {
	values := make(map[types.ScalarAttributeType]int)
	valsSeen := make(map[string]string)

	return func(attrValue string, attrValueType types.ScalarAttributeType) string {
		if vl, ok := valsSeen[attrValue]; ok {
			return vl
		}

		// alias letter
		vl := attrValueType

		// alias number
		values[vl] = values[vl] + 1

		attrValueAlias := fmt.Sprintf(":%s%d", vl, values[vl])

		valsSeen[attrValue] = attrValueAlias

		return attrValueAlias
	}
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
