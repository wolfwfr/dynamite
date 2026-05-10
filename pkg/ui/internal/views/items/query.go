package itemselection

import (
	tea "charm.land/bubbletea/v2"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/util"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

func (m *ItemSelectionPane) enableQueryMode() tea.Cmd {
	if m.queryMode == messages.QueryMode {
		return nil
	}

	m.resetContents()
	m.clearCache()

	m.queryMode = messages.QueryMode
	m.KeyMap.Scan.SetEnabled(true)
	m.KeyMap.ScanParameters.SetEnabled(false)
	m.KeyMap.Query.SetEnabled(false)
	m.KeyMap.QueryParameters.SetEnabled(true)

	switchM := func() tea.Msg {
		return messages.SwitchQueryMode{
			OldMode: m.queryMode,
			NewMode: messages.QueryMode,
		}
	}
	if m.queryParameters.hashKeyValue == "" {
		return tea.Batch(switchM, m.ToggleQueryParametersDialog())
	}
	return tea.Batch(switchM, m.PageNext(true))
}

func (m *ItemSelectionPane) ToggleQueryParametersDialog() tea.Cmd {
	if m.queryMode != messages.QueryMode {
		return nil
	}
	arn := util.IfNotNil(m.selectedTable.TableArn, "")
	sch := m.selectedTable.KeySchema
	hash, rang := primaryKeysFromSchema(sch)
	defs := m.selectedTable.AttributeDefinitions
	defsM := make(map[string]string, len(defs))
	for _, d := range defs {
		defsM[*d.AttributeName] = string(d.AttributeType)
	}

	globalIndices := make([]messages.GlobalSecondaryIndex, len(m.selectedTable.GlobalSecondaryIndexes))
	for i, g := range m.selectedTable.GlobalSecondaryIndexes {
		sch := g.KeySchema
		hash, rang := primaryKeysFromSchema(sch)
		globalIndices[i] = messages.GlobalSecondaryIndex{
			Name:         *g.IndexName,
			HashKey:      hash,
			HashKeyType:  defsM[hash],
			RangeKey:     rang,
			RangeKeyType: defsM[u.IfNotNil(rang, "")],
		}
	}
	localIndices := make([]messages.LocalSecondaryIndex, len(m.selectedTable.LocalSecondaryIndexes))
	for i, l := range m.selectedTable.LocalSecondaryIndexes {
		sch := l.KeySchema
		hash, rang := primaryKeysFromSchema(sch)
		localIndices[i] = messages.LocalSecondaryIndex{
			Name:         *l.IndexName,
			HashKey:      hash,
			HashKeyType:  defsM[hash],
			RangeKey:     *rang,
			RangeKeyType: defsM[*rang],
		}
	}
	index := m.tableIndex.activeIndex
	hashKeyV := m.queryParameters.hashKeyValue
	op := m.queryParameters.rangeKeyOperator
	rangeKeyV1 := m.queryParameters.rangeKeyValue1
	rangeKeyV2 := m.queryParameters.rangeKeyValue2
	orderDesc := m.queryParameters.rangeOrderDescending
	tgl := func() tea.Msg {
		return messages.ToggleQueryParameters{}
	}
	hashType := defsM[hash]
	rangType := defsM[u.IfNotNil(rang, "")]
	init := func() tea.Msg {
		return messages.InitQueryParameters{
			TableARN: arn,
			TableIndex: messages.TableIndex{
				HashKey:      hash,
				HashKeyType:  hashType,
				RangeKey:     rang,
				RangeKeyType: rangType,
			},
			GSI:                  globalIndices,
			LSI:                  localIndices,
			CurrentIndex:         index,
			HashKeyValue:         hashKeyV,
			RangeKeyValue1:       rangeKeyV1,
			RangeKeyValue2:       rangeKeyV2,
			RangeKeyOperator:     op,
			RangeOrderDescending: orderDesc,
		}
	}
	return tea.Batch(tgl, init)
}

func parseRangeKeyOperator(op messages.QueryOperator) types.RangeKeyOperator {
	switch op {
	case messages.Equals:
		return types.RangeEquals
	case messages.Greater:
		return types.RangeGreater
	case messages.GreaterEqual:
		return types.RangeGreaterEqual
	case messages.Less:
		return types.RangeLess
	case messages.LessEqual:
		return types.RangeLessEqual
	case messages.Between:
		return types.RangeBetween
	case messages.BeginsWith:
		return types.RangeBeginsWith
	default:
		return types.RangeEquals
	}
}

func queryPageToPage(page *apitypes.QueryResponse) *messages.Page {
	if page == nil {
		return nil
	}
	return &messages.Page{
		Items:            page.Items,
		LastEvaluatedKey: page.LastEvaluatedKey,
	}
}
