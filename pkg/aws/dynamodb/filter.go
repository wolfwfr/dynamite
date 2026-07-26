package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/util"
)

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
