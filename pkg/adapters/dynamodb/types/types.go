package types

import (
	"image/color"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/lib/styles"
)

type ObjectStyling struct {
	FieldNameColor color.Color
	NumberColor    color.Color
	BoolColor      color.Color
	BytesColor     color.Color
	NULLColor      color.Color
	StringColor    color.Color
	TokenColor     color.Color
	ErrorColor     color.Color
}

// TODO: revise all the duplication between connector & adapter packages
type FilterOperator string

const (
	Noop_F         FilterOperator = ""
	Equals_F       FilterOperator = "equals"
	NotEquals_F    FilterOperator = "not equals"
	GreaterEqual_F FilterOperator = "greater than or equal"
	Greater_F      FilterOperator = "greater than"
	LessEqual_F    FilterOperator = "less than or equal"
	Less_F         FilterOperator = "less than"
	Between_F      FilterOperator = "between"
	Exists_F       FilterOperator = "exists"
	NotExists_F    FilterOperator = "not exists"
	Contains_F     FilterOperator = "contains"
	NotContains_F  FilterOperator = "not contains"
	BeginsWith_F   FilterOperator = "begins with"
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
	// StyledValue  string
	ValueStyling styles.LineStyle
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
	JSONStyled []styles.ObjectStyle
	YAML       []string
	YAMLStyled []styles.ObjectStyle
	Raw        []map[string]types.AttributeValue // TODO: review usefullness
	// TODO: rename to RootKeys
	TableKeys [][]KeyValue // TODO: review: should this be part of items?
}

type ( // FILTER
	FilterExpressionParameters struct {
		AttributePath      string
		AttributeValue1    string
		AttributeValue2    *string
		AttributeValueType types.ScalarAttributeType
		Operator           FilterOperator
	}
)

type ( // SCAN
	ScanParameters struct {
		KeyDetails       []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName        *string                     // optional index-name, queries table if nil
		KeySchema        []types.KeySchemaElement    // keyschema associated with `IndexName` or table
		FilterParameters []FilterExpressionParameters

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
		KeyDetails       []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName        *string                     // optional index-name, queries table if nil
		KeySchema        []types.KeySchemaElement    // keyschema associated with `IndexName` or table
		FilterParameters []FilterExpressionParameters

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
