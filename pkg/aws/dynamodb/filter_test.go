package dynamodb

import (
	"fmt"
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

func TestBuildFilterExpression_NumericalOperators(t *testing.T) {
	t.Run("build-filter-expression should", func(t *testing.T) {
		// convenience resources

		// testcases
		testcases := []struct {
			desc     string
			operator apitypes.FilterOperator
			parsed   string
		}{
			{
				desc:     "correctly parse equal expression",
				operator: apitypes.Equals_F,
				parsed:   "=",
			},
			{
				desc:     "correctly parse not-equal expression",
				operator: apitypes.NotEquals_F,
				parsed:   "<>",
			},
			{
				desc:     "correctly parse gr-eq expression",
				operator: apitypes.GreaterEqual_F,
				parsed:   ">=",
			},
			{
				desc:     "correctly parse gr expression",
				operator: apitypes.Greater_F,
				parsed:   ">",
			},
			{
				desc:     "correctly parse less-eq expression",
				operator: apitypes.LessEqual_F,
				parsed:   "<=",
			},
			{
				desc:     "correctly parse less expression",
				operator: apitypes.Less_F,
				parsed:   "<",
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				params := []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Value",
						AttributeValue1:    "10.5",
						AttributeValueType: types.ScalarAttributeTypeN,
						Operator:           tc.operator,
					},
				}

				expression, expressionNames, expressionValues := buildFilterExpression(params)

				// verify expression
				expExpression := fmt.Sprintf("#V1 %s :N1", tc.parsed)
				require.NotNil(t, expression)
				assert.Equal(t, expExpression, *expression)

				// verify expression-names
				expExpressionNames := map[string]string{"#V1": "Value"}
				require.NotNil(t, expressionNames)
				assert.EqualValues(t, expExpressionNames, expressionNames)

				// verify expression-values
				expExpressionValues := map[string]types.AttributeValue{
					":N1": &types.AttributeValueMemberN{
						Value: "10.5",
					},
				}
				require.NotNil(t, expressionValues)
				assert.EqualValues(t, expExpressionValues, expressionValues)
			})
		}
	})

	t.Run("build-filter-expression should correctly parse between expression", func(t *testing.T) {
		n := "20.5"
		params := []apitypes.FilterExpressionParameters{
			{
				AttributePath:      "Value",
				AttributeValue1:    "10.5",
				AttributeValue2:    &n,
				AttributeValueType: types.ScalarAttributeTypeN,
				Operator:           apitypes.Between_F,
			},
		}

		expression, expressionNames, expressionValues := buildFilterExpression(params)

		// verify expression
		expExpression := "#V1 BETWEEN :N1 AND :N2"
		require.NotNil(t, expression)
		assert.Equal(t, expExpression, *expression)

		// verify expression-names
		expExpressionNames := map[string]string{"#V1": "Value"}
		require.NotNil(t, expressionNames)
		assert.EqualValues(t, expExpressionNames, expressionNames)

		// verify expression-values
		expExpressionValues := map[string]types.AttributeValue{
			":N1": &types.AttributeValueMemberN{
				Value: "10.5",
			},
			":N2": &types.AttributeValueMemberN{
				Value: "20.5",
			},
		}
		require.NotNil(t, expressionValues)
		assert.EqualValues(t, expExpressionValues, expressionValues)
	})
}

func TestBuildFilterExpression_Functions(t *testing.T) {
	t.Run("build-filter-expression should", func(*testing.T) {
		testcases := []struct {
			desc                string
			params              []apitypes.FilterExpressionParameters
			expExpression       string
			expExpressionNames  map[string]string
			expExpressionValues map[string]types.AttributeValue
		}{
			{
				desc: "correctly parse begins_with",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Name",
						AttributeValue1:    "Beginning",
						AttributeValueType: types.ScalarAttributeTypeS,
						Operator:           apitypes.BeginsWith_F,
					},
				},
				expExpression:      "begins_with(#N1,:S1)",
				expExpressionNames: map[string]string{"#N1": "Name"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "Beginning",
					},
				},
			},
			{
				desc: "correctly parse exists",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath: "Name",
						Operator:      apitypes.Exists_F,
					},
				},
				expExpression:      "attribute_exists(#N1)",
				expExpressionNames: map[string]string{"#N1": "Name"},
			},
			{
				desc: "correctly parse not_exists",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath: "Name",
						Operator:      apitypes.NotExists_F,
					},
				},
				expExpression:      "attribute_not_exists(#N1)",
				expExpressionNames: map[string]string{"#N1": "Name"},
			},
			{
				desc: "correctly parse contains",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Name",
						AttributeValue1:    "Elem",
						AttributeValueType: types.ScalarAttributeTypeS,
						Operator:           apitypes.Contains_F,
					},
				},
				expExpression:      "contains(#N1,:S1)",
				expExpressionNames: map[string]string{"#N1": "Name"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "Elem",
					},
				},
			},
			{
				desc: "correctly parse not_contains",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath:      "Name",
						AttributeValue1:    "Elem",
						AttributeValueType: types.ScalarAttributeTypeS,
						Operator:           apitypes.NotContains_F,
					},
				},
				expExpression:      "NOT contains(#N1,:S1)",
				expExpressionNames: map[string]string{"#N1": "Name"},
				expExpressionValues: map[string]types.AttributeValue{
					":S1": &types.AttributeValueMemberS{
						Value: "Elem",
					},
				},
			},
			{
				desc: "correctly parses a path",
				params: []apitypes.FilterExpressionParameters{
					{
						AttributePath: "Parent.Child",
						Operator:      apitypes.Exists_F,
					},
				},
				expExpression:      "attribute_exists(#P1.#C1)",
				expExpressionNames: map[string]string{"#P1": "Parent", "#C1": "Child"},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.desc, func(t *testing.T) {
				expression, expressionNames, expressionValues := buildFilterExpression(tc.params)

				// verify expression
				require.NotNil(t, expression)
				assert.Equal(t, tc.expExpression, *expression)

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
