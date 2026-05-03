package messages

import (
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

type View int
type ItemsQueryMode int

const (
	Table_selection View = iota
	Item_selection
)
const (
	ScanMode ItemsQueryMode = iota
	QueryMode
)

type QueryOperator string

const (
	Noop         QueryOperator = ""
	Equals       QueryOperator = "equals"
	GreaterEqual QueryOperator = "greater than or equal"
	Greater      QueryOperator = "greater than"
	LessEqual    QueryOperator = "less than or equal"
	Less         QueryOperator = "less than"
	Between      QueryOperator = "between"
	BeginsWith   QueryOperator = "begins with"
)

type TableIndex struct {
	HashKey     string
	HashKeyType string

	RangeKey     *string
	RangeKeyType string
}

type GlobalSecondaryIndex struct {
	Name string

	HashKey     string
	HashKeyType string

	RangeKey     *string
	RangeKeyType string
}

type LocalSecondaryIndex struct {
	Name string

	HashKey     string
	HashKeyType string

	RangeKey     string
	RangeKeyType string
}

type SwitchView struct {
	OldView View
	NewView View
}

type SelectTable struct {
	TableName    string
	TableDetails apitypes.DescribeTableResponse
}

type ZoomToggleItemSelectionPane struct{}
type ZoomToggleItemDetailsPane struct{}
type ZoomToggleTableSelectionPane struct{}
type ZoomToggleTableDetailsPane struct{}

type PreviewItem struct {
	Item string
}

type TableDetails struct {
	Details *apitypes.DescribeTableResponse
}

type ToggleJSONYAML struct{}

type Page struct {
	Items            apitypes.Items
	LastEvaluatedKey map[string]dynamodbtypes.AttributeValue
}

type PageReady struct {
	Table    apitypes.DescribeTableResponse
	Index    *string
	Response *Page
	Err      error
}

type TablePageReady struct {
	Tables        []string
	PaginationKey *string
	Err           error
	Region        string
}

type ToggleHelp struct{}
type ToggleRegions struct{}

type ToggleColumnVisibility struct{}
type ToggleColumnSorting struct{}
type ToggleScanParameters struct{}
type ToggleQueryParameters struct{}

type InitColumnVisibility struct {
	TableARN   string
	AllColumns []string // matching by index
	Visible    []bool   // matching by index
}

type InitColumnSorting struct {
	TableARN   string
	AllColumns []string // matching by index
	SortingOn  string
	Ascending  bool // if false, descending
}

type InitScanParameters struct {
	TableARN     string
	TableIndex   TableIndex
	GSI          []GlobalSecondaryIndex
	LSI          []LocalSecondaryIndex
	CurrentIndex *string
}

type InitQueryParameters struct {
	TableARN         string
	TableIndex       TableIndex
	GSI              []GlobalSecondaryIndex
	LSI              []LocalSecondaryIndex
	CurrentIndex     *string
	HashKeyValue     string
	RangeKeyValue1   *string
	RangeKeyValue2   *string // used for BETWEEN operator
	RangeKeyOperator QueryOperator
}

type ColumnVisibilityUpdate struct {
	TableARN   string
	AllColumns []string // matching by index
	Visible    []bool   // matching by index
}

type ColumnSortingUpdate struct {
	TableARN   string
	AllColumns []string // matching by index
	SortingOn  string
	Ascending  bool // if false, descending
}

type ScanIndexChanged struct {
	TableARN  string
	IndexName string // empty == table index
}

type QueryParametersChanged struct {
	TableARN         string
	IndexName        string // empty == table index
	HashKeyValue     string
	RangeKeyValue1   *string
	RangeKeyValue2   *string // used for BETWEEN operator
	RangeKeyOperator QueryOperator
}

type ColumnSortingReset struct {
	TableARN string
}

type SwitchRegion struct {
	OldRegion string
	NewRegion string
}

type SwitchQueryMode struct {
	OldMode ItemsQueryMode
	NewMode ItemsQueryMode
}

type CopyItem struct{}
