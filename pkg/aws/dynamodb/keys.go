package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

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

func formatQueryKeys(params apitypes.QueryParameters) (hashkey string, rangekey *string, KeyConditionExpression string, ExpressionAttributeValues map[string]types.AttributeValue) {
	hash, rang := parsePrimaryKeys(params.KeySchema)

	hashkey = *hash
	rangekey = rang

	var hashType *types.ScalarAttributeType
	var rangType *types.ScalarAttributeType

	// obtain key-attribute types (string, number, boolean)
	for _, d := range params.KeyDetails {
		if *d.AttributeName == *hash {
			hashType = &d.AttributeType
			if rang == nil || rangType != nil {
				break
			}
		}
		if rang != nil && *d.AttributeName == *rang {
			rangType = &d.AttributeType
			if hashType != nil {
				break
			}
		}
	}

	// prepare the key-condition-expression
	var keys string
	{
		keys = fmt.Sprintf("%s = %s", *hash, ":partitionkeyval")
		if rang != nil {
			keys = fmt.Sprintf("%s AND %s = %s", keys, *rang, ":sortkeyval")
		}
	}
	KeyConditionExpression = keys

	var hashAttrVal types.AttributeValue
	var rangAttrVal types.AttributeValue

	// prepare the expression-attribute-values
	{
		hashAttrVal = matchScalarAttrType(*hashType, params.HashKeyValue)
		if rangType != nil {
			rangAttrVal = matchScalarAttrType(*rangType, params.RangeKeyValue)
		}
	}

	ExpressionAttributeValues = map[string]types.AttributeValue{
		":partitionkeyval": hashAttrVal,
	}
	if rangAttrVal != nil {
		ExpressionAttributeValues[":sortkeyval"] = rangAttrVal
	}
	return
}

func matchScalarAttrType(typ types.ScalarAttributeType, val string) types.AttributeValue {
	var v types.AttributeValue
	switch typ {
	case types.ScalarAttributeTypeS:
		v = &types.AttributeValueMemberS{Value: val}
	case types.ScalarAttributeTypeN:
		v = &types.AttributeValueMemberN{Value: val}
	case types.ScalarAttributeTypeB:
		v = &types.AttributeValueMemberB{Value: []byte(val)}
	}
	return v
}
