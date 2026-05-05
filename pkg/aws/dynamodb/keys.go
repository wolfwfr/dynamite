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

// formatQueryKeys parses query parameters and prepares inputs for a dynamodb
// query.
func formatQueryKeys(params apitypes.QueryParameters) (hashkey string, rangekey *string, expression string, expressionValues map[string]types.AttributeValue, expressionNames map[string]string, err error) {
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
		keys = fmt.Sprintf("#hashkey = %s", ":partitionkeyval")
		if rang != nil && params.RangeKeyValue1 != nil {
			newKeys := fmt.Sprintf("%s AND #rangekey %s %s", keys, parseRangeOperator(params.RangeKeyOperator), ":sortkeyval")
			if params.RangeKeyOperator == apitypes.RangeBetween {
				if params.RangeKeyValue2 == nil {
					err = fmt.Errorf("cannot apply 'between' range operator without 2 appliccable values")
					return
				}
				newKeys = fmt.Sprintf("%s AND #rangekey %s %s AND %s", keys, parseRangeOperator(params.RangeKeyOperator), ":sortkeyval", ":sortkeyval2")
			}
			if params.RangeKeyOperator == apitypes.RangeBeginsWith {
				// begins_with ( sortKeyName , :sortkeyval )
				newKeys = fmt.Sprintf("%s AND %s ( #rangekey , %s )", keys, parseRangeOperator(params.RangeKeyOperator), ":sortkeyval")
			}
			keys = newKeys
		}
	}
	expression = keys

	var hashAttrVal types.AttributeValue
	var rangAttrVal1 types.AttributeValue
	var rangAttrVal2 types.AttributeValue

	// prepare the expression-attribute-values
	hashAttrVal = matchScalarAttrType(*hashType, params.HashKeyValue)
	if rangType != nil && params.RangeKeyValue1 != nil {
		rangAttrVal1 = matchScalarAttrType(*rangType, *params.RangeKeyValue1)
	}
	if rangType != nil && params.RangeKeyValue2 != nil {
		rangAttrVal2 = matchScalarAttrType(*rangType, *params.RangeKeyValue2)
	}

	// using expression-attr-names prevents name collisions with dynamodb
	// reserved words, see: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/ReservedWords.html

	// set expression-names
	expressionNames = map[string]string{
		"#hashkey": *hash,
	}
	if rang != nil && rangAttrVal1 != nil {
		expressionNames["#rangekey"] = *rang
	}

	// set expression-values
	expressionValues = map[string]types.AttributeValue{
		":partitionkeyval": hashAttrVal,
	}
	if rangAttrVal1 != nil {
		expressionValues[":sortkeyval"] = rangAttrVal1
	}
	if rangAttrVal2 != nil && params.RangeKeyOperator == apitypes.RangeBetween {
		expressionValues[":sortkeyval2"] = rangAttrVal2
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
