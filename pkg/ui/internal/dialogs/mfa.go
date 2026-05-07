package dialogs

import (
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	appconfig "github.com/wolfwfr/dynamite/pkg"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/messages"
	"github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
	commonstyles "github.com/wolfwfr/dynamite/pkg/ui/internal/styles"
)

// the MFA dialog requests an MFA token for AWS credits
type MFA struct {
	styles mfaStyles

	keyMap mfaKeyMap

	defaultDialogHeight int
	defaultDialogWidth  int
	window              struct {
		width  int
		height int
	}
	dialog struct {
		width  int
		height int
	}

	input textinput.Model

	help help.Model

	credsC chan<- appconfig.CredentialsResponse
}

type mfaKeyMap struct {
	enter key.Binding
	close key.Binding
}

func (h mfaKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{h.close, h.enter}
}

func (h mfaKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{h.close, h.enter}}
}

type mfaStyles struct {
	dialogStyle lipgloss.Style

	title    lipgloss.Style
	desc     lipgloss.Style
	inputBox lipgloss.Style
	helpLine lipgloss.Style
}

func newMFAStyles() mfaStyles {
	s := mfaStyles{}
	s.dialogStyle = commonstyles.DialogStyle
	s.title = lipgloss.NewStyle().Padding(1, 0, 1, 0)
	s.desc = lipgloss.NewStyle().Padding(1, 0, 1, 0).Foreground(styles.SubtleColour)
	s.helpLine = lipgloss.NewStyle().Padding(1, 4, 0, 4)
	s.inputBox = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(commonstyles.DialogFocusColour)
	return s
}

func NewMFADialog(credsC chan<- appconfig.CredentialsResponse) *MFA {
	d := &MFA{}

	{ // keymap
		d.keyMap = mfaKeyMap{
			enter: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "accept input"),
			),
			close: key.NewBinding(
				key.WithKeys("esc"),
				key.WithHelp("esc", "cancel input"),
			),
		}
	}

	{ // output
		d.credsC = credsC
	}

	{ //sizes
		d.defaultDialogHeight = 10
		d.defaultDialogWidth = 40

		d.dialog.width = d.defaultDialogWidth
		d.dialog.height = d.defaultDialogHeight

		d.window.width = 150
		d.window.height = 100
	}

	{ // styles
		d.styles = newMFAStyles()
	}

	{ // user input
		input := textinput.New()
		input.Prompt = "Token > "
		input.CharLimit = 6

		d.input = input
	}

	{ // help
		d.help = help.New()
	}

	return d
}

func (m *MFA) Init() tea.Cmd {
	return nil
}

func (m *MFA) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keyMap.close):
			return m.cancel()
		case key.Matches(msg, m.keyMap.enter):
			return m.accept()
		}
	}
	switch msg := msg.(type) {
	case messages.MFAFocus:
		return m.input.Focus()
	case tea.WindowSizeMsg:
		m.applySize(msg)
		return nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return cmd
	}
	return nil
}

func (m *MFA) cancel() tea.Cmd {
	m.credsC <- appconfig.CredentialsResponse{
		Token: "",
		Error: fmt.Errorf("user canceled"),
	}
	return func() tea.Msg {
		return messages.CloseMFADialog{}
	}
}

func (m *MFA) accept() tea.Cmd {
	m.credsC <- appconfig.CredentialsResponse{
		Token: m.input.Value(),
		Error: nil,
	}
	return func() tea.Msg {
		return messages.CloseMFADialog{}
	}
}

func (m *MFA) applySize(msg tea.WindowSizeMsg) {
	m.window.width = msg.Width
	m.window.height = msg.Height
	m.updateSize()
}

func (m *MFA) updateSize() {
	s := newMFAStyles()

	s.dialogStyle.Width(m.dialog.width)
	s.dialogStyle.Height(m.dialog.height)
	s.inputBox.Width(m.dialog.width / 2)

	m.styles = s

	m.input.SetWidth(20)
}

func (m *MFA) View() string {
	title := "AWS MFA Token"
	desc := "Your AWS profile calls for an MFA token,\nplease provide it below or provide\nenvironment variable credentials."
	return m.styles.dialogStyle.Render(
		lipgloss.JoinVertical(lipgloss.Center,
			m.styles.title.Render(title),
			m.styles.desc.Render(desc),
			m.styles.inputBox.Render(m.input.View()),
			m.styles.helpLine.Render(m.help.View(m.keyMap)),
		),
	)
}
