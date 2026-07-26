package dynamodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

func TestBuildFilterExpression(t *testing.T) {
	t.Run("build-filter-expression should", func(t *testing.T) {
		// convenience resources

		// testcases
		testcases := []struct {
			desc                string
			params              []apitypes.FilterExpressionParameters
			expExpression       string
			expNilExpression    bool
			expExpressionNames  map[string]string
			expExpressionValues map[string]types.AttributeValue
		}{
			{
				desc:                "return only nil values when there are no filters",
				params:              make([]apitypes.FilterExpressionParameters, 0),
				expNilExpression:    true,
				expExpressionNames:  nil,
				expExpressionValues: nil,
			},
			{
				desc: "asign value alias based on type",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Name",
						AttributeValue1:    "Value",
						AttributeValueType: "S",
						Operator:           apitypes.Equals_F,
					},
					{
						AttributePath:      "Name",
						AttributeValue1:    "12.5",
						AttributeValueType: "N",
						Operator:           apitypes.Equals_F,
					},
					{
						AttributePath:      "Name",
						AttributeValue1:    "true",
						AttributeValueType: "B",
						Operator:           apitypes.Equals_F,
					},
				},
				expExpression:      "#N1 = :S1 AND #N1 = :N1 AND #N1 = :B1",
				expExpressionNames: map[string]string{"#N1": "Name"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "Value",
					},
					":N1": &types.AttributeValueMemberN{
						Value: "12.5",
					},
					":B1": &types.AttributeValueMemberBOOL{
						Value: true,
					},
				},
			},
			{
				desc: "use unique expression-names & -values for repeated attribute-names & -values",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Name",
						AttributeValue1:    "Value",
						AttributeValueType: "S",
						Operator:           apitypes.Equals_F,
					},
					{
						AttributePath:      "Name",
						AttributeValue1:    "Value",
						AttributeValueType: "S",
						Operator:           apitypes.NotEquals_F,
					},
				},
				expExpression:      "#N1 = :S1 AND #N1 <> :S1",
				expExpressionNames: map[string]string{"#N1": "Name"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "Value",
					},
				},
			},
			{
				desc: "increment alias number for both names and values",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "NameX",
						AttributeValue1:    "ValueX",
						AttributeValueType: "S",
						Operator:           apitypes.Equals_F,
					},
					{
						AttributePath:      "NameY",
						AttributeValue1:    "ValueY",
						AttributeValueType: "S",
						Operator:           apitypes.Equals_F,
					},
				},
				expExpression:      "#N1 = :S1 AND #N2 = :S2",
				expExpressionNames: map[string]string{"#N1": "NameX", "#N2": "NameY"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "ValueX",
					},
					":S2": &types.AttributeValueMemberS{
						Value: "ValueY",
					},
				},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				expression, expressionNames, expressionValues := buildFilterExpression(tc.params)
				if tc.expNilExpression {
					assert.Nil(t, expression)
				} else {
					require.NotNil(t, expression)
					assert.Equal(t, tc.expExpression, *expression)
				}
				if tc.expExpressionNames == nil {
					assert.Nil(t, expressionNames)
				} else {
					require.NotNil(t, expressionNames)
					assert.EqualValues(t, tc.expExpressionNames, expressionNames)
				}
				if tc.expExpressionValues == nil {
					assert.Nil(t, expressionValues)
				} else {
					require.NotNil(t, expressionValues)
					assert.EqualValues(t, tc.expExpressionValues, expressionValues)
				}
			})
		}
	})
}
