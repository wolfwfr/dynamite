package itemselection

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/search"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
)

//go:generate mockgen -source=$GOFILE -destination=./mocks/gen.go -package=mocks
type dynamodbClient interface {
	ScanTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.ScanParameters) (*apitypes.ScanResponse, error)
	QueryTable(client *dynamodb.Client, ctx context.Context, table string, params apitypes.QueryParameters) (*apitypes.QueryResponse, error)
}

type itemsTable interface {
	GetColumns() []table.Column
	GetRows() []table.Row
	GetVirtualRows() []table.Row
	GetVisualRows() []table.Row
	GetKeyMap() *table.KeyMap
	GetDynamicColumnWidth() bool
	GetSelectedRow() *table.Row
	GetSelectedItem() (*itemstable.Item, int)
	GetAllowedOptions() viewoptions.Check
	GetViewOptionsState() viewoptions.ViewOptions

	AddItems(items apitypes.Items, hasRangeKey bool)

	Reset()
	ResetSearch()
	ResetColumnVisibility()
	ResetColumnSorting()

	SetColumnSorting(cols []string, sortingOn string, ascending bool) bool
	SetColumnVisibility(cols []string, visible []bool) bool
	SetSearchEnable() bool
	SetSearchResults(col string, results []search.FilteredItem) bool
	SetDynamicColumnWidth(b bool)

	PaginationEligible() bool
	UpdateSize(height, width int)

	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	View() string
}
