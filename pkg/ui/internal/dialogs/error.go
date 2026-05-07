package dialogs

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/google/uuid"

	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

type errorDialogStyles struct {
	errorStyle  lipgloss.Style
	dialogStyle lipgloss.Style
}

// the Welcome dialog depicts a welcome message
type ErrorDialog struct {
	id string

	error error

	styles errorDialogStyles

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

type Option func(e *ErrorDialog)

func WithDuration(d time.Duration) Option {
	return func(e *ErrorDialog) {
		e.duration = d
	}
}

func WithClosingKey(k key.Binding) Option {
	return func(e *ErrorDialog) {
		e.close = &k
	}
}

func NewErrorDialog(err error, opts ...Option) *ErrorDialog {
	d := &ErrorDialog{
		id:       uuid.New().String()[:6],
		error:    err,
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

func (m *ErrorDialog) newStyles() {
	s := errorDialogStyles{}
	s.dialogStyle = commonstyles.DialogStyle.Align(lipgloss.Left, lipgloss.Center)
	s.errorStyle = lipgloss.NewStyle().PaddingBottom(1)
	m.styles = s
}

// to initialise ticking
func (m *ErrorDialog) Tick() tea.Cmd {
	return m.tick()
}

func (m *ErrorDialog) tick() tea.Cmd {
	id := m.id
	return func() tea.Msg {
		<-m.tracker.ticker.C
		return messages.ErrorTick{ID: id}
	}
}

func (m *ErrorDialog) ID() string {
	return m.id
}

func (m *ErrorDialog) Init() tea.Cmd {
	return nil
}

func (m *ErrorDialog) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applySize(msg)
	case messages.ErrorTick:
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

func (m *ErrorDialog) onTick() tea.Cmd {
	return tea.Batch(m.tick(), m.checkExpiry())
}

func (m *ErrorDialog) toggleDialog() tea.Cmd {
	id := m.id
	return func() tea.Msg {
		return messages.ErrorExpired{
			ID: id,
		}
	}
}

func (m *ErrorDialog) checkExpiry() tea.Cmd {
	if m.tracker.finished {
		m.tracker.ticker.Stop()
		return m.toggleDialog()
	}
	return nil
}

func (m *ErrorDialog) applySize(msg tea.WindowSizeMsg) {
	m.window.width = msg.Width
	m.window.height = msg.Height
	m.updateSize()
}

func (m *ErrorDialog) updateSize() {
	m.styles.dialogStyle = m.styles.dialogStyle.
		MaxHeight(m.window.height - 10).
		MaxWidth(m.window.width - 10).
		Width(m.dialog.width)
}

func (m *ErrorDialog) View() string {
	return m.styles.dialogStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.errorStyle.Render(m.error.Error()),
		m.renderProgress(),
	))
}

func (m *ErrorDialog) renderProgress() string {
	w := m.dialog.width - 4

	p := m.getProgression()
	m.tracker.finished = p >= 1

	b := int(p * float64(w))

	s := strings.Builder{}

	for range b {
		s.WriteString("─") // TODO: replace
	}

	return s.String()
}

// returns fraction in [0, 1]
func (m *ErrorDialog) getProgression() float64 {
	return min(1, float64(time.Since(m.tracker.start))/float64(m.duration))
}
