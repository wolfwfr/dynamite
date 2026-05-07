package types

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RangeKeyOperator string

const (
	RangeEquals       RangeKeyOperator = "equals"
	RangeGreater      RangeKeyOperator = "greater than"
	RangeGreaterEqual RangeKeyOperator = "greater than or equals"
	RangeLess         RangeKeyOperator = "less than"
	RangeLessEqual    RangeKeyOperator = "less than or equals"
	RangeBetween      RangeKeyOperator = "between"
	RangeBeginsWith   RangeKeyOperator = "begins with"
)

type KeyValue struct {
	Key   string
	Value string
}

type ( // LIST TABLES
	ListTablesRequest struct {
		LastEvaluatedTableName *string
		Limit                  *int32
	}
	ListTablesResponse struct {
		TableNames             []string
		LastEvaluatedTableName *string
	}
)

type ( // DESCRIBE TABLE
	// for more information on
	// `github.com/aws/aws-sdk-go-v2/service/dynamodb/types/types.go:TableDescription`
	DescribeTableResponse struct {
		AttributeDefinitions      []types.AttributeDefinition
		BillingModeSummary        *types.BillingModeSummary
		CreationDateTime          *time.Time
		DeletionProtectionEnabled *bool
		GlobalSecondaryIndexes    []types.GlobalSecondaryIndexDescription
		ItemCount                 *int64
		KeySchema                 []types.KeySchemaElement
		LocalSecondaryIndexes     []types.LocalSecondaryIndexDescription
		OnDemandThroughput        *types.OnDemandThroughput
		ProvisionedThroughput     *types.ProvisionedThroughputDescription
		SSEDescription            *types.SSEDescription
		TableArn                  *string
		TableClassSummary         *types.TableClassSummary
		TableId                   *string
		TableName                 *string
		TableSizeBytes            *int64
	}
)

type Items struct {
	JSON       []string
	JSONStyled []string
	YAML       []string
	YAMLStyled []string
	Raw        []map[string]types.AttributeValue // TODO: review usefullness
	TableKeys  [][]KeyValue                      // TODO: review: should this be part of items?
}

type ( // SCAN
	ScanParameters struct {
		KeyDetails []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName  *string                     // optional index-name, queries table if nil
		KeySchema  []types.KeySchemaElement    // keyschema associated with `IndexName` or table

		Limit            int
		LastEvaluatedKey map[string]types.AttributeValue
	}
	ScanResponse struct {
		Items            Items
		LastEvaluatedKey map[string]types.AttributeValue
	}
)

type ( // QUERY
	QueryParameters struct {
		KeyDetails []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName  *string                     // optional index-name, queries table if nil
		KeySchema  []types.KeySchemaElement    // keyschema associated with `IndexName` or table

		HashKeyValue     string  // required
		RangeKeyValue1   *string // optional
		RangeKeyValue2   *string // used for BETWEEN operator
		RangeKeyOperator RangeKeyOperator

		Limit            int
		LastEvaluatedKey map[string]types.AttributeValue
		Descending       bool // default to ascending
	}
	QueryResponse struct {
		Items            Items
		LastEvaluatedKey map[string]types.AttributeValue
	}
)
