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

func ParseFilterOperator(op apitypes.FilterOperator) string {
	switch op {
	case apitypes.Equals_F:
		return "="
	case apitypes.NotEquals_F:
		return "<>"
	case apitypes.GreaterEqual_F:
		return ">="
	case apitypes.Greater_F:
		return ">"
	case apitypes.LessEqual_F:
		return "<="
	case apitypes.Less_F:
		return "<"
	case apitypes.Between_F:
		return "BETWEEN"
	case apitypes.Exists_F:
		panic("not implemented yet!")
	case apitypes.NotExists_F:
		panic("not implemented yet!")
	case apitypes.Contains_F:
		panic("not implemented yet!")
	case apitypes.NotContains_F:
		panic("not implemented yet!")
	case apitypes.BeginsWith_F:
		panic("not implemented yet!")
	default:
		return ""
	}
}
