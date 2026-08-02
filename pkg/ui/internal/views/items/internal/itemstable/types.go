package itemstable

import (
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"image/color"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/wolfwfr/dynamite/lib/styles"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/common"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/components/table"
	view "github.com/wolfwfr/dynamite/pkg/ui/internal/views/items/internal/itemstable/viewoptions"
)

const ItemIndexMetaKey = "item_index"

type TableStyles struct {
	SelectedBackground    color.Color
	SearchMatchBackground color.Color
}

// ItemsTable is a custom table implementation that is specialised for dynamo-db
// items, including styling, and view modelations (e.g. sorting, and search)
// TODO: consider renaming to 'Model'
type ItemsTable struct {
	// access to the tables current contents
	table *table.Model

	// render-cache caches row-fields rendered by the table's field-delegate
	renderCache map[string]string

	// styles
	styles TableStyles

	// dynamo-db-Items including JSON/YAML render & styling instructions
	Items apitypes.Items

	viewOptions view.ViewOptions

	// ColumnAttributes represents a unique set of dynamo-db item keys that
	// exhaustively cover all keys in the currently paged set of items
	ColumnAttributes []ColumnAttributes
}

type ColumnAttributes struct {
	Title string
	Type  common.DynamoDBAttributeType
}

// TODO: refactor dynamodb.Items and add single Item
type Item struct {
	JSON       string
	JSONStyled styles.ObjectStyle
	YAML       string
	YAMLStyled styles.ObjectStyle
	Raw        map[string]dynamodbtypes.AttributeValue
	TableKeys  []apitypes.KeyValue
}

// EnrichedField defines the field-type that populates a table-row.
type EnrichedField struct {
	RawValue string
	Style    *styles.LineStyle
}

// Value implements the matching table.Field interface function
func (f EnrichedField) Value() string {
	return f.RawValue
}
