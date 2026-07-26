package itemselection

import (
	tea "charm.land/bubbletea/v2"

	apitypes "github.com/wolfwfr/dynamite/pkg/ui/internal/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

func (m *ItemSelectionPane) ChangeFilterParameters(msg messages.FilterParametersChanged) tea.Cmd {
	if u.IfNotNil(m.selectedTable.TableArn, "") != msg.TableARN { // expired
		return nil
	}

	params := make([]apitypes.FilterExpressionParameters, 0, len(msg.State))

	for _, st := range msg.State {
		if st.AttrPath == "" {
			continue
		}
		p := apitypes.FilterExpressionParameters{
			AttributePath:      st.AttrPath,
			AttributeValue1:    u.IfNotNil(st.AttrValue1, ""), // allowed nil on 'exists' or 'not_exists' operator
			AttributeValue2:    st.AttrValue2,
			AttributeValueType: st.AttrType,
			Operator:           apitypes.FilterOperator(st.FilterOperator),
		}
		params = append(params, p)
	}

	// cancel paging, and refresh table contents
	m.resetContents()

	mode := m.queryMode
	if mode == messages.QueryMode {
		m.filterParameters.query = params
	}
	if mode == messages.ScanMode {
		m.filterParameters.scan = params
	}

	return m.PageNext(true)
}

// toggle filter-parameters diagram
func (m *ItemSelectionPane) ToggleFilterParametersDialog() tea.Cmd {
	arn := u.IfNotNil(m.selectedTable.TableArn, "")
	toggle := func() tea.Msg {
		return messages.ToggleFilterParameters{}
	}
	state := func() tea.Msg {
		msg := messages.InitFilterParameters{}
		msg.TableARN = arn
		msg.TableAttrDefinitions = m.selectedTable.AttributeDefinitions
		params := m.filterParameters.query
		if m.queryMode == messages.ScanMode {
			params = m.filterParameters.scan
		}
		st := make([]messages.FilterState, len(params))
		for i, p := range params {
			st[i] = messages.FilterState{
				AttrPath:       p.AttributePath,
				AttrType:       p.AttributeValueType,
				AttrValue1:     &p.AttributeValue1,
				AttrValue2:     p.AttributeValue2,
				FilterOperator: messages.FilterOperator(p.Operator),
			}
		}
		msg.State = st
		return msg
	}
	return tea.Batch(toggle, state)
}
