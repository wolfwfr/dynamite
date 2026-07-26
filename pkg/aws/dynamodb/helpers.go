package dynamodb

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

func ToAttrValue(value string, attrType types.ScalarAttributeType) types.AttributeValue {
	switch attrType {
	case types.ScalarAttributeTypeS:
		return &types.AttributeValueMemberS{
			Value: value,
		}
	case types.ScalarAttributeTypeN:
		return &types.AttributeValueMemberN{
			Value: value,
		}
	case types.ScalarAttributeTypeB:
		b, _ := strconv.ParseBool(value) // TODO: error-checking?
		return &types.AttributeValueMemberBOOL{
			Value: b,
		}
	default:
		return &types.AttributeValueMemberS{
			Value: value,
		}
	}
}

// see: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html
const (
	filterEqual        string = "="
	filterNotEqual     string = "<>"
	filterGreaterEqual string = ">="
	filterGreater      string = ">"
	filterLessEqual    string = "<="
	filterLess         string = "<"
	filterBetween      string = "BETWEEN"
	filterExists       string = "attribute_exists"
	filterNotExists    string = "attribute_not_exists"
	filterBeginsWith   string = "begins_with"
	filterContains     string = "contains"
)

func ParseFilterComparator(op apitypes.FilterOperator) string {
	switch op {
	case apitypes.Equals_F:
		return filterEqual
	case apitypes.NotEquals_F:
		return filterNotEqual
	case apitypes.GreaterEqual_F:
		return filterGreaterEqual
	case apitypes.Greater_F:
		return filterGreater
	case apitypes.LessEqual_F:
		return filterLessEqual
	case apitypes.Less_F:
		return filterLess
	case apitypes.Between_F:
		return filterBetween
	default:
		return ""
	}
}

func filterComparators() []apitypes.FilterOperator {
	return []apitypes.FilterOperator{
		apitypes.Equals_F,
		apitypes.NotEquals_F,
		apitypes.GreaterEqual_F,
		apitypes.Greater_F,
		apitypes.LessEqual_F,
		apitypes.Less_F,
	}
}
