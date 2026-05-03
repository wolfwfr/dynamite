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

func formatQueryKeys(params apitypes.QueryParameters) (hashkey string, rangekey *string, KeyConditionExpression string, ExpressionAttributeValues map[string]types.AttributeValue, err error) {
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
		if rang != nil && params.RangeKeyValue1 != nil {
			newKeys := fmt.Sprintf("%s AND %s %s %s", keys, *rang, parseRangeOperator(params.RangeKeyOperator), ":sortkeyval")
			if params.RangeKeyOperator == apitypes.RangeBetween {
				if params.RangeKeyValue2 == nil {
					err = fmt.Errorf("cannot apply 'between' range operator without 2 appliccable values")
					return
				}
				newKeys = fmt.Sprintf("%s AND %s %s %s AND %s", keys, *rang, parseRangeOperator(params.RangeKeyOperator), ":sortkeyval", ":sortkeyval2")
			}
			if params.RangeKeyOperator == apitypes.RangeBeginsWith {
				// begins_with ( sortKeyName , :sortkeyval )
				newKeys = fmt.Sprintf("%s AND %s ( %s , %s )", keys, parseRangeOperator(params.RangeKeyOperator), *rang, ":sortkeyval")
			}
			keys = newKeys
		}
	}
	KeyConditionExpression = keys

	var hashAttrVal types.AttributeValue
	var rangAttrVal1 types.AttributeValue
	var rangAttrVal2 types.AttributeValue

	// prepare the expression-attribute-values
	{
		hashAttrVal = matchScalarAttrType(*hashType, params.HashKeyValue)
		if rangType != nil && params.RangeKeyValue1 != nil {
			rangAttrVal1 = matchScalarAttrType(*rangType, *params.RangeKeyValue1)
		}
		if rangType != nil && params.RangeKeyValue2 != nil {
			rangAttrVal2 = matchScalarAttrType(*rangType, *params.RangeKeyValue2)
		}
	}

	ExpressionAttributeValues = map[string]types.AttributeValue{
		":partitionkeyval": hashAttrVal,
	}
	if rangAttrVal1 != nil {
		ExpressionAttributeValues[":sortkeyval"] = rangAttrVal1
	}
	if rangAttrVal2 != nil && params.RangeKeyOperator == apitypes.RangeBetween {
		ExpressionAttributeValues[":sortkeyval2"] = rangAttrVal2
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

func parseRangeOperator(op apitypes.RangeKeyOperator) string {
	switch op {
	case apitypes.RangeEquals:
		return "="
	case apitypes.RangeGreater:
		return ">"
	case apitypes.RangeGreaterEqual:
		return ">="
	case apitypes.RangeLess:
		return "<"
	case apitypes.RangeLessEqual:
		return "<="
	case apitypes.RangeBetween:
		return "between"
	case apitypes.RangeBeginsWith:
		return "begins_with"
	default:
		return "="
	}
}
