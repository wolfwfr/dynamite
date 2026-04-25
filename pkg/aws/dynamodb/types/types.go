package types

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type KeyValue struct {
	Key   string
	Value string
}

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

type ( // SCAN
	ScanParameters struct {
		KeyDetails []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName  *string                     // optional index-name, queries table if nil
		KeySchema  []types.KeySchemaElement    // keyschema associated with `IndexName` or table

		Limit int
		// LastEvaluatedKey map[string]types.AttributeValue
	}
	ScanResponse struct {
		ItemsJSON []string
		ItemsYAML []string
		ItemsRaw  []map[string]types.AttributeValue // TODO: review usefullness
		TableKeys [][]KeyValue
		// LastEvaluatedKey map[string]types.AttributeValue
	}
)

type ( // QUERY
	QueryParameters struct {
		KeyDetails []types.AttributeDefinition // table attribute-definitions, describing table & index key attribute types
		IndexName  *string                     // optional index-name, queries table if nil
		KeySchema  []types.KeySchemaElement    // keyschema associated with `IndexName` or table

		HashKeyValue  string // required
		RangeKeyValue string // optional

		// Limit int
		// LastEvaluatedKey map[string]types.AttributeValue
	}
	QueryResponse struct {
		ItemsJSON []string
		ItemsYAML []string
		ItemsRaw  []map[string]types.AttributeValue // TODO: review usefullness
		TableKeys [][]KeyValue

		// LastEvaluatedKey map[string]types.AttributeValue
	}
)
