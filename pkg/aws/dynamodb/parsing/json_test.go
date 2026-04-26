package parsing

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func genericTestItem() map[string]types.AttributeValue {
	item := map[string]types.AttributeValue{}
	item["string-key"] = &types.AttributeValueMemberS{
		Value: "string-value",
	}
	item["string-set-key"] = &types.AttributeValueMemberSS{
		Value: []string{"string-set-value-1", "string-set-value-2"},
	}
	item["number-key"] = &types.AttributeValueMemberN{
		Value: "100.5",
	}
	item["number-set-key"] = &types.AttributeValueMemberNS{
		Value: []string{"52.5", "35"},
	}
	item["byte-key"] = &types.AttributeValueMemberB{
		Value: []byte("byte-value"),
	}
	item["byte-set-key"] = &types.AttributeValueMemberBS{
		Value: [][]byte{[]byte("byte-set-value-1"), []byte("byte-set-value-2")},
	}
	item["list-key-1"] = &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: "list-1-value-1"},
			&types.AttributeValueMemberS{Value: "list-1-value-2"},
		},
	}
	item["list-key-2"] = &types.AttributeValueMemberL{
		Value: []types.AttributeValue{
			&types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"list-2-value-1-field-1": &types.AttributeValueMemberS{Value: "map-1-value-1"},
					"list-2-value-1-field-2": &types.AttributeValueMemberS{Value: "map-1-value-2"},
				},
			},
			&types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"list-2-value-2-field-1": &types.AttributeValueMemberS{Value: "map-2-value-1"},
					"list-2-value-2-field-2": &types.AttributeValueMemberS{Value: "map-2-value-2"},
				},
			},
		},
	}
	item["map-key"] = &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"map-value-field-1": &types.AttributeValueMemberS{Value: "map-value-field-1-value"},
			"map-value-field-2": &types.AttributeValueMemberS{Value: "map-value-field-2-value"},
		},
	}
	item["null-true-key"] = &types.AttributeValueMemberNULL{
		Value: true,
	}
	item["null-false-key"] = &types.AttributeValueMemberNULL{
		Value: false,
	}
	item["bool-true-key"] = &types.AttributeValueMemberBOOL{
		Value: true,
	}
	item["bool-false-key"] = &types.AttributeValueMemberBOOL{
		Value: false,
	}
	item["empty-set-key"] = &types.AttributeValueMemberSS{}
	item["empty-map-key"] = &types.AttributeValueMemberM{}

	return item
}

func genericTestItemJSON() string {
	tabsize = 2
	return `{
  "string-key": "string-value",
  "bool-false-key": false,
  "bool-true-key": true,
  "byte-key": <bytes>(len=10),
  "byte-set-key": [
    <bytes>(len=16),
    <bytes>(len=16)
  ],
  "empty-map-key": {},
  "empty-set-key": [],
  "list-key-1": [
    "list-1-value-1",
    "list-1-value-2"
  ],
  "list-key-2": [
    {
      "list-2-value-1-field-1": "map-1-value-1",
      "list-2-value-1-field-2": "map-1-value-2"
    },
    {
      "list-2-value-2-field-1": "map-2-value-1",
      "list-2-value-2-field-2": "map-2-value-2"
    }
  ],
  "map-key": {
    "map-value-field-1": "map-value-field-1-value",
    "map-value-field-2": "map-value-field-2-value"
  },
  "null-false-key": NOT NULL,
  "null-true-key": NULL,
  "number-key": 100.5,
  "number-set-key": [
    52.5,
    35
  ],
  "string-set-key": [
    "string-set-value-1",
    "string-set-value-2"
  ]
}`
}

func TestJSONParsing(t *testing.T) {
	item := genericTestItem()
	exp := genericTestItemJSON()
	res := ParseItemToJSON(item, "string-key", nil)
	assert.EqualValues(t, exp, res)
}
