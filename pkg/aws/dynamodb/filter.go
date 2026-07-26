package dynamodb

import (
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/util"
)

// NOTE: not supported at this time:
// - NOT
// - OR
// - IN
// - size
// - attribute_type()
// see: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Expressions.OperatorsAndFunctions.html

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
		// TODO: use strings.Builder
		var s string
		if p.AttributePath == "" || p.Operator == apitypes.Noop_F || p.Operator == apitypes.Between_F && util.IfNotNil(p.AttributeValue2, "") == "" {
			continue
		}

		var (
			pathAlias      string
			attrValueAlias string
		)

		pathEle := strings.Split(p.AttributePath, ".")
		pathAliases := make([]string, len(pathEle))
		for i, e := range pathEle {
			attrNameAlias := nameAliasConstructor(e)
			exprNames[attrNameAlias] = e
			pathAliases[i] = attrNameAlias
		}
		pathAlias = strings.Join(pathAliases, ".")

		if p.AttributeValue1 != "" {
			attrValueAlias = valueAliasConstructor(p.AttributeValue1, p.AttributeValueType)
			exprVals[attrValueAlias] = ToAttrValue(p.AttributeValue1, p.AttributeValueType)
		}

		if i > 0 {
			s = " AND " // TODO: consider supporting 'OR' operator
		}

		switch {
		case slices.Contains(filterComparators(), p.Operator):
			s = fmt.Sprintf("%s%s %s %s", s, pathAlias, ParseFilterComparator(p.Operator), attrValueAlias)
		case p.Operator == apitypes.Between_F:
			attrVal2 := util.IfNotNil(p.AttributeValue2, "")
			attrValue2Alias := valueAliasConstructor(attrVal2, p.AttributeValueType)
			exprVals[attrValue2Alias] = ToAttrValue(attrVal2, p.AttributeValueType)
			s = fmt.Sprintf("%s%s %s %s AND %s", s, pathAlias, ParseFilterComparator(p.Operator), attrValueAlias, attrValue2Alias)
		case p.Operator == apitypes.Exists_F:
			s = fmt.Sprintf("%s%s(%s)", s, filterExists, pathAlias)
		case p.Operator == apitypes.NotExists_F:
			s = fmt.Sprintf("%s%s(%s)", s, filterNotExists, pathAlias)
		case p.Operator == apitypes.Contains_F:
			s = fmt.Sprintf("%s%s(%s,%s)", s, filterContains, pathAlias, attrValueAlias)
		case p.Operator == apitypes.NotContains_F:
			s = fmt.Sprintf("%sNOT %s(%s,%s)", s, filterContains, pathAlias, attrValueAlias)
		case p.Operator == apitypes.BeginsWith_F:
			s = fmt.Sprintf("%s%s(%s,%s)", s, filterBeginsWith, pathAlias, attrValueAlias)
		default:
			// TODO: debug logging or error
			continue
		}
		exprr = fmt.Sprintf("%s%s", exprr, s)
	}
	expr = &exprr

	// ensure nil if empty; required by dynamodb API
	if expr != nil && len(*expr) == 0 {
		expr = nil
	}
	if len(exprNames) == 0 {
		exprNames = nil
	}
	if len(exprVals) == 0 {
		exprVals = nil
	}

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
