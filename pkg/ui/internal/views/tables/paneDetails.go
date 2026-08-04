package tableselection

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	apitypes "github.com/wolfwfr/dynamite/pkg/adapters/dynamodb/types"
	"github.com/wolfwfr/dynamite/pkg/logging"
	"github.com/wolfwfr/dynamite/pkg/theme"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/views/util/keymaps"
	u "github.com/wolfwfr/dynamite/pkg/util"
)

type detailsPane struct {
	// shared config
	config *appconfig.Config

	// logger
	logger *slog.Logger

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

	styles detailsStyles

	content viewport.Model
}

type detailsStyles struct {
	headerStyle     lipgloss.Style
	fieldNameStyle  lipgloss.Style
	fieldValueStyle lipgloss.Style
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
		logger:  config.Logger.With(slog.String(logging.ViewKey, Log_TablesView), slog.String(logging.PaneKey, "details-pane")),
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

	p.updateStyles()

	return p
}

func (m *detailsPane) updateStyles() {
	m.logger.Debug("updating styles")
	m.styles = detailsStyles{
		headerStyle:     lipgloss.NewStyle().Bold(true).Background(theme.TableDetailsHeaderBg).Foreground(theme.TableDetailsHeaderFg).PaddingBottom(1),
		fieldNameStyle:  lipgloss.NewStyle().Background(theme.TableDetailsFieldNameBg).Foreground(theme.TableDetailsFieldNameFg),   //.Bold(true),
		fieldValueStyle: lipgloss.NewStyle().Background(theme.TableDetailsFieldValueBg).Foreground(theme.TableDetailsFieldValueFg), //.Bold(true),
	}
}

func (m *detailsPane) cleanSlate() {
	m.err = nil
}

func (m *detailsPane) Init() tea.Cmd {
	m.logger.Info("initialising...")
	m.cleanSlate()
	m.logger.Info("initialisation complete")
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
		m.content.SetYOffset(0)
		m.content.SetContent(renderDetails(msg.Details, m.styles))
		return nil
	case tea.BackgroundColorMsg:
		m.updateStyles()
		return nil
	}

	m.content, cmd = m.content.Update(msg)
	return
}

func renderDetails(details *apitypes.DescribeTableResponse, styles detailsStyles) string {
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

	name := u.IfNotNil(details.TableName, "")
	arn := u.IfNotNil(details.TableArn, "")
	id := u.IfNotNil(details.TableId, "")

	header := styles.headerStyle.Render
	field := styles.fieldNameStyle.Render
	value := styles.fieldValueStyle.Render
	spf := fmt.Sprintf

	s := strings.Builder{}
	fmt.Fprintf(&s, "%s\n", header("GENERAL"))
	fmt.Fprintf(&s, "%s:   %s\n", field("Table name"), value(name))
	fmt.Fprintf(&s, "%s:    %s\n", field("Table ARN"), value(arn))
	fmt.Fprintf(&s, "%s:     %s\n", field("Table ID"), value(id))
	fmt.Fprintf(&s, "\n")
	if details.TableClassSummary != nil {
		fmt.Fprintf(&s, "%s:  %s\n", field("Table Class"), value(string(details.TableClassSummary.TableClass)))
		fmt.Fprintf(&s, "\n")
	}
	fmt.Fprintf(&s, "%s:   %s\n", field("Created At"), value(details.CreationDateTime.Format(time.RFC1123Z)))
	fmt.Fprintf(&s, "\n")
	fmt.Fprintf(&s, "%s\n", header("COUNT"))
	fmt.Fprintf(&s, "%s:                   %s\n", field("Table Item Count"), value(spf("%d", *details.ItemCount)))
	fmt.Fprintf(&s, "%s:  %s\n", field("Global Secondary Index Item Count"), value(spf("%d", globalIdxItemCount)))
	fmt.Fprintf(&s, "%s:   %s\n", field("Local Secondary Index Item Count"), value(spf("%d", localIdxItemCount)))
	fmt.Fprintf(&s, "\n")
	fmt.Fprintf(&s, "%s\n", header("SIZE"))
	fmt.Fprintf(&s, "%s:                   %s\n", field("Total Size"), value(formatBytes(totalSize)))
	fmt.Fprintf(&s, "%s:                   %s\n", field("Table Size"), value(formatBytes(*details.TableSizeBytes)))
	fmt.Fprintf(&s, "%s:  %s\n", field("Global Secondary Index Size"), value(formatBytes(globalIdxSize)))
	fmt.Fprintf(&s, "%s:   %s\n", field("Local Secondary Index Size"), value(formatBytes(localIdxSize)))
	fmt.Fprintf(&s, "\n")
	fmt.Fprintf(&s, "%s\n", header("TABLE KEYS"))
	hash, rang := primaryKeysFromSchema(details.KeySchema)
	fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "", details.AttributeDefinitions, styles))
	fmt.Fprintf(&s, "\n")
	if len(details.GlobalSecondaryIndexes) > 0 {
		fmt.Fprintf(&s, "%s\n", header("GLOBAL SECONDARY INDICES"))
		for i, idx := range details.GlobalSecondaryIndexes {
			fmt.Fprintf(&s, "%s: %s\n", field("Index Name"), value(*idx.IndexName))
			fmt.Fprintf(&s, "%s:  %s\n", field("Index ARN"), value(*idx.IndexArn))
			fmt.Fprintf(&s, "\n")
			hash, rang := primaryKeysFromSchema(idx.KeySchema)
			fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "  ", details.AttributeDefinitions, styles))
			if i != len(details.GlobalSecondaryIndexes)-1 {
				fmt.Fprintf(&s, "\n")
			}
		}
		fmt.Fprintf(&s, "\n")
	}
	if len(details.LocalSecondaryIndexes) > 0 {
		fmt.Fprintf(&s, "%s\n", header("LOCAL SECONDARY INDICES"))
		for i, idx := range details.LocalSecondaryIndexes {
			fmt.Fprintf(&s, "%s: %s\n", field("Index Name"), value(*idx.IndexName))
			fmt.Fprintf(&s, "%s:  %s\n", field("Index ARN"), value(*idx.IndexArn))
			fmt.Fprintf(&s, "\n")
			hash, rang := primaryKeysFromSchema(idx.KeySchema)
			fmt.Fprintf(&s, "%s", formatKeys(hash, rang, "  ", details.AttributeDefinitions, styles))
			if i != len(details.LocalSecondaryIndexes)-1 {
				fmt.Fprintf(&s, "\n")
			}
		}
		fmt.Fprintf(&s, "\n")
	}
	fmt.Fprintf(&s, "%s\n", header("SECURITY"))
	fmt.Fprintf(&s, "%s: %s\n", field("Deletion Protection Enabled"), value(spf("%t", *details.DeletionProtectionEnabled)))
	fmt.Fprintf(&s, "\n")
	if details.BillingModeSummary != nil {
		fmt.Fprintf(&s, "%s\n", header("BILLING"))
		fmt.Fprintf(&s, "%s: %s\n", field("Billing Mode"), value(string(details.BillingModeSummary.BillingMode)))
		fmt.Fprintf(&s, "\n")
	}
	if details.ProvisionedThroughput != nil {
		fmt.Fprintf(&s, "%s\n", header("THROUGHPUT PROVISIONED"))
		fmt.Fprintf(&s, "%s:   %s\n", field("Read Capacity Units"), value(spf("%d", *details.ProvisionedThroughput.ReadCapacityUnits)))
		fmt.Fprintf(&s, "%s:  %s\n", field("Write Capacity Units"), value(spf("%d", *details.ProvisionedThroughput.WriteCapacityUnits)))
		fmt.Fprintf(&s, "\n")
	}
	if details.OnDemandThroughput != nil {
		fmt.Fprintf(&s, "%s\n", header("THROUGHPUT ON DEMAND"))
		fmt.Fprintf(&s, "%s:   %s\n", field("Max Read Capacity Units"), value(spf("%d", *details.OnDemandThroughput.MaxReadRequestUnits)))
		fmt.Fprintf(&s, "%s:  %s\n", field("Max Write Capacity Units"), value(spf("%d", *details.OnDemandThroughput.MaxWriteRequestUnits)))
		fmt.Fprintf(&s, "\n")
	}
	return s.String()
}

func formatKeys(hash string, rang *string, indentation string, attrDef []dynamodbtypes.AttributeDefinition, styles detailsStyles) string {
	field := styles.fieldNameStyle.Render
	value := styles.fieldValueStyle.Render
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
			rangAttr = u.ToPtr(string(d.AttributeType))
			if hashAttr != "" {
				break
			}
		}
	}
	hashfmt := fmt.Sprintf("%s%s:  %s\n", indentation, field(fmt.Sprintf("Hash Key  (%s)", hashAttr)), value(hash))
	if rang == nil {
		return hashfmt
	}
	return fmt.Sprintf("%s%s%s:  %s\n", hashfmt, indentation, field(fmt.Sprintf("Range Key (%s)", *rangAttr)), value(*rang))
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
	if m.err != nil {
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
