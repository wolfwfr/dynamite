package messages

import "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"

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
	TableDetails types.DescribeTableResponse
}

type ZoomToggleItemSelectionPane struct{}
type ZoomToggleItemDetailsPane struct{}
type ZoomToggleTableSelectionPane struct{}
type ZoomToggleTableDetailsPane struct{}

type PreviewItem struct {
	Item string
}
