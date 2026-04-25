package parsing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genericTestItemYAML() string {
	tabsize = 2
	return `string-key: "string-value"
bool-false-key: false
bool-true-key: true
byte-key: byte-value
byte-set-key: 
  - byte-set-value-1
  - byte-set-value-2
empty-map-key: 
empty-set-key: 
list-key-1: 
  - "list-1-value-1"
  - "list-1-value-2"
list-key-2: 
  - list-2-value-1-field-1: "map-1-value-1"
    list-2-value-1-field-2: "map-1-value-2"
  - list-2-value-2-field-1: "map-2-value-1"
    list-2-value-2-field-2: "map-2-value-2"
map-key: 
  map-value-field-1: "map-value-field-1-value"
  map-value-field-2: "map-value-field-2-value"
null-false-key: NOT NULL
null-true-key: NULL
number-key: 100.5
number-set-key: 
  - 52.5
  - 35
string-set-key: 
  - "string-set-value-1"
  - "string-set-value-2"
`
}

func TestYAMLParsing(t *testing.T) {
	tabsize = 2
	item := genericTestItem()
	res := ParseItemToYAML(item, "string-key", nil)
	exp := genericTestItemYAML()
	assert.EqualValues(t, exp, res)
	fmt.Print(res)
}
