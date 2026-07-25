package dynamodb

import (
	cncrtypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
)

func convertFilterParameters(p []types.FilterExpressionParameters) []cncrtypes.FilterExpressionParameters {
	res := make([]cncrtypes.FilterExpressionParameters, len(p))
	for i, pp := range p {
		res[i] = cncrtypes.FilterExpressionParameters{
			AttributeName:      pp.AttributeName,
			AttributeValue1:    pp.AttributeValue1,
			AttributeValue2:    pp.AttributeValue2,
			AttributeValueType: pp.AttributeValueType,
			Operator:           cncrtypes.FilterOperator(pp.Operator),
		}
	}
	return res
}
