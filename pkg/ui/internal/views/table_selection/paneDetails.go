package tableselection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	apitypes "github.com/wolfwfr/dynamite/pkg/aws/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/keymaps"
)

type detailsPane struct {
	// shared config
	config *appconfig.Config

	// errorText
	err error

	// pane's view window
	window struct {
		width  int
		height int
	}

	// key map
	KeyMap *DetailsPaneKeyMap

	// Additional Keys
	AddKeyMap keymaps.AdditionalKeys

	content viewport.Model
}

type detailsPaneOption func(p *detailsPane)

func withDetailsPaneKeys(keys keymaps.AdditionalKeys) detailsPaneOption {
	return func(t *detailsPane) {
		t.AddKeyMap = keys
	}
}

func newDetailsPane(ctx context.Context, config *appconfig.Config, opts ...detailsPaneOption) *detailsPane {
	step := 5
	c := viewport.New(viewport.WithHeight(20)) // content
	c.SoftWrap = false
	c.SetHorizontalStep(step)
	c.KeyMap.Left.SetHelp("←/h", "left")
	c.KeyMap.Right.SetHelp("→/l", "right")
	p := &detailsPane{
		config:  config,
		content: c,
		KeyMap:  DefaultDetailsKeyMap(),
	}
	for _, o := range opts {
		o(p)
	}

	if !keymaps.UniqueKeyMaps(p.KeyMap.ShortHelp(), p.AddKeyMap.Bindings()) {
		panic("overlapping keymaps!")
	}
	return p
}

func (m *detailsPane) cleanSlate() {
	m.err = nil
}

func (m *detailsPane) Init() tea.Cmd {
	m.cleanSlate()
	return nil
}

func (m *detailsPane) Update(msg tea.Msg) (cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Zoom):
			return m.Zoom()
		default:
			if match, call := m.AddKeyMap.Matches(msg); match {
				return call
			}
		}
	case messages.TableDetails:
		m.content.SetContent(renderDetails(msg.Details))
		return nil
	}

	m.content, cmd = m.content.Update(msg)
	return
}

func renderDetails(details *apitypes.DescribeTableResponse) string {
	if details == nil {
		return ""
	}
	totalSize := *details.TableSizeBytes
	globalIdxSize := int64(0)
	globalIdxItemCount := int64(0)
	localIdxSize := int64(0)
	localIdxItemCount := int64(0)
	for _, i := range details.GlobalSecondaryIndexes {
		totalSize += *i.IndexSizeBytes
		globalIdxSize += *i.IndexSizeBytes
		globalIdxItemCount += *i.ItemCount
	}
	for _, i := range details.LocalSecondaryIndexes {
		totalSize += *i.IndexSizeBytes
		localIdxSize += *i.IndexSizeBytes
		localIdxItemCount += *i.ItemCount
	}

	name := ternarySafe(details.TableName, "", details.TableName != nil)
	arn := ternarySafe(details.TableArn, "", details.TableArn != nil)
	id := ternarySafe(details.TableId, "", details.TableId != nil)

	s := strings.Builder{}
	fmt.Fprintf(&s, "----------------------------------------------------\n")
	fmt.Fprintf(&s, "[GENERAL]\n")
	fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	fmt.Fprintf(&s, "Table name:   %s\n", name)
	fmt.Fprintf(&s, "Table ARN:    %s\n", arn)
	fmt.Fprintf(&s, "Table ID:     %s\n", id)
	fmt.Fprintf(&s, "\n")
	if details.TableClassSummary != nil {
		fmt.Fprintf(&s, "Table Class:  %s\n", details.TableClassSummary.TableClass)
		fmt.Fprintf(&s, "\n")
	}
	fmt.Fprintf(&s, "Created At:   %s\n", details.CreationDateTime.Format(time.RFC1123Z))
	fmt.Fprintf(&s, "\n----------------------------------------------------\n")
	fmt.Fprintf(&s, "[COUNT]\n")
	fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	fmt.Fprintf(&s, "Table Item Count:                   %d\n", *details.ItemCount)
	fmt.Fprintf(&s, "Global Secondary Index Item Count:  %d\n", globalIdxItemCount)
	fmt.Fprintf(&s, "Local Secondary Index Itemm Count:  %d\n", localIdxItemCount)
	fmt.Fprintf(&s, "\n----------------------------------------------------\n")
	fmt.Fprintf(&s, "[SIZE]\n")
	fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	fmt.Fprintf(&s, "Total Size:                   %s\n", formatBytes(totalSize))
	fmt.Fprintf(&s, "Table Size:                   %s\n", formatBytes(*details.TableSizeBytes))
	fmt.Fprintf(&s, "Global Secondary Index Size:  %s\n", formatBytes(globalIdxSize))
	fmt.Fprintf(&s, "Local Secondary Index Size:   %s\n", formatBytes(localIdxSize))
	fmt.Fprintf(&s, "\n----------------------------------------------------\n")
	fmt.Fprintf(&s, "[TABLE KEYS]\n")
	fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	hash, rang := primaryKeysFromSchema(details.KeySchema)
	fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "", details.AttributeDefinitions))
	if len(details.GlobalSecondaryIndexes) > 0 {
		fmt.Fprintf(&s, "\n----------------------------------------------------\n")
		fmt.Fprintf(&s, "[GLOBAL SECONDARY INDICES]\n")
		fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	}
	for i, idx := range details.GlobalSecondaryIndexes {
		fmt.Fprintf(&s, "Index Name: %s\n", *idx.IndexName)
		fmt.Fprintf(&s, "Index ARN:  %s\n", *idx.IndexArn)
		fmt.Fprintf(&s, "\n")
		hash, rang := primaryKeysFromSchema(idx.KeySchema)
		fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "  ", details.AttributeDefinitions))
		if i != len(details.GlobalSecondaryIndexes)-1 {
			fmt.Fprintf(&s, "\n")
		}
	}
	if len(details.LocalSecondaryIndexes) > 0 {
		fmt.Fprintf(&s, "\n----------------------------------------------------\n")
		fmt.Fprintf(&s, "[LOCAL SECONDARY INDICES]\n")
		fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	}
	for i, idx := range details.LocalSecondaryIndexes {
		fmt.Fprintf(&s, "Index Name: %s\n", *idx.IndexName)
		fmt.Fprintf(&s, "Index ARN:  %s\n", *idx.IndexArn)
		fmt.Fprintf(&s, "\n")
		hash, rang := primaryKeysFromSchema(idx.KeySchema)
		fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "  ", details.AttributeDefinitions))
		if i != len(details.LocalSecondaryIndexes)-1 {
			fmt.Fprintf(&s, "\n")
		}
	}
	fmt.Fprintf(&s, "\n----------------------------------------------------\n")
	fmt.Fprintf(&s, "[SECURITY]\n")
	fmt.Fprintf(&s, "----------------------------------------------------\n\n")
	fmt.Fprintf(&s, "Deletion Protection Enabled: %t\n", *details.DeletionProtectionEnabled)
	if details.BillingModeSummary != nil {
		fmt.Fprintf(&s, "\n----------------------------------------------------\n")
		fmt.Fprintf(&s, "[BILLING]\n")
		fmt.Fprintf(&s, "----------------------------------------------------\n\n")
		fmt.Fprintf(&s, "Billing Mode: %s\n", details.BillingModeSummary.BillingMode)
	}
	if details.ProvisionedThroughput != nil {
		fmt.Fprintf(&s, "\n----------------------------------------------------\n")
		fmt.Fprintf(&s, "[THROUGHPUT PROVISIONED]\n")
		fmt.Fprintf(&s, "----------------------------------------------------\n\n")
		fmt.Fprintf(&s, "Read Capacity Units:   %d\n", *details.ProvisionedThroughput.ReadCapacityUnits)
		fmt.Fprintf(&s, "Write Capacity Units:  %d\n", *details.ProvisionedThroughput.WriteCapacityUnits)
	}
	if details.OnDemandThroughput != nil {
		fmt.Fprintf(&s, "\n----------------------------------------------------\n")
		fmt.Fprintf(&s, "[THROUGHPUT ON DEMAND]\n")
		fmt.Fprintf(&s, "----------------------------------------------------\n\n")
		fmt.Fprintf(&s, "Max Read Capacity Units:   %d\n", *details.OnDemandThroughput.MaxReadRequestUnits)
		fmt.Fprintf(&s, "Max Write Capacity Units:  %d\n", *details.OnDemandThroughput.MaxWriteRequestUnits)
	}
	return s.String()
}

func formatKeys(hash string, rang *string, indentation string, attrDef []dynamodbtypes.AttributeDefinition) string {
	var hashAttr string
	var rangAttr *string
	for _, d := range attrDef {
		if hash == *d.AttributeName {
			hashAttr = string(d.AttributeType)
			if rangAttr != nil || rang == nil {
				break
			}
		}
		if rang != nil && *rang == *d.AttributeName {
			rangAttr = toPtr(string(d.AttributeType))
			if hashAttr != "" {
				break
			}
		}
	}
	hashfmt := fmt.Sprintf("%sHash Key  (%s):  %s\n", indentation, hashAttr, hash)
	if rang == nil {
		return hashfmt
	}
	return fmt.Sprintf("%s%sRange Key (%s):  %s\n", hashfmt, indentation, *rangAttr, *rang)
}

func formatBytes(bytes int64) string {
	bytesF := float64(bytes)
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	var i int
	for i < len(sizes)-1 {
		if bytesF < 1000 {
			break
		}
		i++
		bytesF = bytesF / 1000
	}
	return fmt.Sprintf("%.2f %s", bytesF, sizes[i])
}

func (m *detailsPane) Zoom() tea.Cmd {
	return func() tea.Msg {
		return messages.ZoomToggleTableDetailsPane{}
	}
}

func (m *detailsPane) applySize(height, width int) {
	// m.content.applySize(m.window.height-2-3, m.window.width/2-4)
	m.window.height = height
	m.window.width = width
	m.content.SetHeight(height)
	m.content.SetWidth(width)
}

func (m *detailsPane) View() string {
	if m.err != nil { // TODO: formatting
		return m.err.Error()
	}
	return m.content.View()
}

func primaryKeysFromSchema(s []dynamodbtypes.KeySchemaElement) (hash string, rang *string) {
	for _, e := range s {
		if e.KeyType == dynamodbtypes.KeyTypeHash {
			hash = *e.AttributeName
		} else {
			rang = e.AttributeName
		}
	}
	return
}

func toPtr[T any](t T) *T {
	return &t
}

// with appropriate condition, it escapes nil-pointers
func ternarySafe[T any](first *T, second T, cond bool) T {
	if cond {
		return *first
	}
	return second
}
