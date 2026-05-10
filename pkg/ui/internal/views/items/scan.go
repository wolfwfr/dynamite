package itemselection

import (
	tea "charm.land/bubbletea/v2"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/util"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

// enableScanMode immediately returns when already enabled and forc == false. It
// calls pageNext with initialisation when enabling.
func (m *ItemSelectionPane) enableScanMode(force bool) tea.Cmd {
	if m.queryMode == messages.ScanMode && !force {
		return nil
	}

	m.resetContents()
	m.clearCache()

	m.queryMode = messages.ScanMode
	m.tableIndex.activeIndex = m.scanParameters.index
	m.KeyMap.Query.SetEnabled(true)
	m.KeyMap.QueryParameters.SetEnabled(false)
	m.KeyMap.Scan.SetEnabled(false)
	m.KeyMap.ScanParameters.SetEnabled(true)

	switchM := func() tea.Msg {
		return messages.SwitchQueryMode{
			OldMode: m.queryMode,
			NewMode: messages.ScanMode,
		}
	}
	return tea.Batch(switchM, m.PageNext(true))
}

func (m *ItemSelectionPane) ToggleScanParametersDialog() tea.Cmd {
	if m.queryMode != messages.ScanMode {
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
	tgl := func() tea.Msg {
		return messages.ToggleScanParameters{}
	}
	index := m.tableIndex.activeIndex
	init := func() tea.Msg {
		return messages.InitScanParameters{
			TableARN: arn,
			TableIndex: messages.TableIndex{
				HashKey:      hash,
				HashKeyType:  defsM[hash],
				RangeKey:     rang,
				RangeKeyType: defsM[u.IfNotNil(rang, "")],
			},
			GSI:          globalIndices,
			LSI:          localIndices,
			CurrentIndex: index,
		}
	}
	return tea.Batch(tgl, init)
}

func scanPageToPage(page *apitypes.ScanResponse) *messages.Page {
	if page == nil {
		return nil
	}
	return &messages.Page{
		Items:            page.Items,
		LastEvaluatedKey: page.LastEvaluatedKey,
	}
}
