package messages

import (
	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
)

type View int

const (
	Table_selection View = iota
	Item_selection
)

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
	Details apitypes.DescribeTableResponse
}

type ToggleJSONYAML struct{}

type ScanPageReady struct {
	Table    apitypes.DescribeTableResponse
	Index    *string
	Response apitypes.ScanResponse
}

type ToggleHelp struct{}
type ToggleRegions struct{}

type SwitchRegion struct {
	OldRegion string
	NewRegion string
}
