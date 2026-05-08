package dialogs

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/google/uuid"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

type notificationDialogStyles struct {
	messageStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	dividerStyle  lipgloss.Style
	progressStyle lipgloss.Style
	dialogStyle   lipgloss.Style
}

// the Welcome dialog depicts a welcome message
type NotificationDialog struct {
	id string

	err error
	msg string

	styles notificationDialogStyles

	duration time.Duration

	close *key.Binding

	tracker struct {
		start    time.Time
		ticker   *time.Ticker
		finished bool
	}

	defaultDialogHeight int
	defaultDialogWidth  int

	window struct {
		width  int
		height int
	}

	dialog struct {
		width  int
		height int
	}
}

type Option func(e *NotificationDialog)

func WithDuration(d time.Duration) Option {
	return func(e *NotificationDialog) {
		e.duration = d
	}
}

func WithClosingKey(k key.Binding) Option {
	return func(e *NotificationDialog) {
		e.close = &k
	}
}

func NewNotificationDialog(msg string, err error, opts ...Option) *NotificationDialog {
	d := &NotificationDialog{
		id:       uuid.New().String()[:6],
		err:      err,
		msg:      msg,
		duration: 5 * time.Second,

		defaultDialogHeight: 46,
		defaultDialogWidth:  55,
	}

	d.tracker.start = time.Now()

	d.dialog.width = d.defaultDialogWidth
	d.dialog.height = d.defaultDialogHeight

	d.window.width = 150
	d.window.height = 100

	for _, o := range opts {
		o(d)
	}

	d.tracker.ticker = time.NewTicker(50 * time.Millisecond)

	d.newStyles()
	d.updateSize()

	return d
}

func (m *NotificationDialog) newStyles() {
	s := notificationDialogStyles{}
	s.dialogStyle = commonstyles.DialogStyle.Align(lipgloss.Left, lipgloss.Center)
	s.messageStyle = lipgloss.NewStyle()
	s.errorStyle = lipgloss.NewStyle()
	s.dividerStyle = lipgloss.NewStyle().Foreground(styles.SubtleColour3)
	s.progressStyle = lipgloss.NewStyle().PaddingTop(1)
	m.styles = s
}

// to initialise ticking
func (m *NotificationDialog) Tick() tea.Cmd {
	return m.tick()
}

func (m *NotificationDialog) tick() tea.Cmd {
	id := m.id
	return func() tea.Msg {
		<-m.tracker.ticker.C
		return messages.NotificationTick{ID: id}
	}
}

func (m *NotificationDialog) ID() string {
	return m.id
}

func (m *NotificationDialog) Init() tea.Cmd {
	return nil
}

func (m *NotificationDialog) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applySize(msg)
	case messages.NotificationTick:
		if msg.ID == m.id {
			return m.onTick()
		}
		return nil
	case tea.KeyPressMsg:
		if m.close != nil && key.Matches(msg, *m.close) {
			cmd = m.toggleDialog()
		}
	}

	return tea.Batch(cmd, m.checkExpiry())
}

func (m *NotificationDialog) onTick() tea.Cmd {
	return tea.Batch(m.tick(), m.checkExpiry())
}

func (m *NotificationDialog) toggleDialog() tea.Cmd {
	id := m.id
	return func() tea.Msg {
		return messages.NotificationExpired{
			ID: id,
		}
	}
}

func (m *NotificationDialog) checkExpiry() tea.Cmd {
	if m.tracker.finished {
		m.tracker.ticker.Stop()
		return m.toggleDialog()
	}
	return nil
}

func (m *NotificationDialog) applySize(msg tea.WindowSizeMsg) {
	m.window.width = msg.Width
	m.window.height = msg.Height
	m.updateSize()
}

func (m *NotificationDialog) updateSize() {
	m.styles.dialogStyle = m.styles.dialogStyle.
		MaxHeight(m.window.height - 10).
		MaxWidth(m.window.width - 10).
		Width(m.dialog.width)
}

func (m *NotificationDialog) View() string {
	var main string
	if m.msg != "" {
		main = m.styles.messageStyle.Render(m.msg)
	}
	if m.err != nil && m.msg != "" {
		main = lipgloss.JoinVertical(lipgloss.Left, main, m.styles.dividerStyle.Render(m.renderDivider()))
	}
	if m.err != nil {
		main = lipgloss.JoinVertical(lipgloss.Left,
			main,
			m.styles.errorStyle.Render(m.err.Error()),
		)
	}
	return m.styles.dialogStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		main,
		m.styles.progressStyle.Render(m.renderProgress()),
	))
}

func (m *NotificationDialog) renderProgress() string {
	w := m.dialog.width - 4

	p := m.getProgression()
	m.tracker.finished = p >= 1

	b := int(p * float64(w))

	s := strings.Builder{}

	for range b {
		s.WriteString("─")
	}

	return s.String()
}

func (m *NotificationDialog) renderDivider() string {
	w := m.dialog.width - 4

	s := strings.Builder{}

	for range w {
		s.WriteString("─")
	}

	return s.String()
}

// returns fraction in [0, 1]
func (m *NotificationDialog) getProgression() float64 {
	return min(1, float64(time.Since(m.tracker.start))/float64(m.duration))
}
