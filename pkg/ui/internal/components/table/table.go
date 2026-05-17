// Package table is a customized implementation of the simple table implementation
// at "charm.land/bubbles/v2/table".
package table

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// New creates a new model for the table widget.
func New(opts ...Option) *Model {
	step := 5
	h := viewport.New(viewport.WithHeight(2)) // header
	h.SoftWrap = false                        // disable text-wrap and allow horizontal scroll
	h.SetHorizontalStep(step)
	c := viewport.New(viewport.WithHeight(20)) // content
	c.SoftWrap = h.SoftWrap
	c.SetHorizontalStep(step)
	m := Model{
		cursor:  0,
		content: c, //nolint:mnd
		header:  h,

		dynCols: true,

		keyMap: DefaultKeyMap(),
		help:   help.New(),
		styles: DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	m.UpdateContent()

	return &m
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

// SetStyles sets the table styles.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateContent()
}

// Update is the Bubble Tea update loop.
func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if !m.focus {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.keyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.keyMap.ScrollRight):
			m.ScrollRight(1)
		case key.Matches(msg, m.keyMap.ScrollLeft):
			m.ScrollLeft(1)
		case key.Matches(msg, m.keyMap.ShiftRight):
			m.ScrollRight(m.content.Width() / 4)
		case key.Matches(msg, m.keyMap.ShiftLeft):
			m.ScrollLeft(m.content.Width() / 4)
		case key.Matches(msg, m.keyMap.PageUp):
			m.MoveUp(m.content.Height())
		case key.Matches(msg, m.keyMap.PageDown):
			m.MoveDown(m.content.Height())
		case key.Matches(msg, m.keyMap.HalfPageUp):
			m.MoveUp(m.content.Height() / 2) //nolint:mnd
		case key.Matches(msg, m.keyMap.HalfPageDown):
			m.MoveDown(m.content.Height() / 2) //nolint:mnd
		case key.Matches(msg, m.keyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.keyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.keyMap.GotoLeft):
			m.GotoLeft()
		case key.Matches(msg, m.keyMap.GotoRight):
			m.GotoRight()
		}
	}

	return nil
}

// View renders the component.
func (m *Model) View() string {
	return m.header.View() + "\n" + m.content.View()
}

// HelpView is a helper method for rendering the help menu from the keymap.
// Note that this view is not rendered by default and you must call it
// manually in your application, where applicable.
func (m *Model) HelpView() string {
	return m.help.View(m.keyMap)
}

func (m *Model) MoveContentBoundaries(n int) {
	rows := m.VisualRows()

	m.start = clamp(m.start+n, 0, max(0, len(rows)-m.content.Height()))
	m.end = min(m.start+m.content.Height(), len(rows))
}

func (m *Model) updateContentHeight() {
	if m.Height() < 1 || m.cursor < 0 {
		return
	}
	oldLen := m.content.TotalLineCount()
	newLen := m.content.Height()
	diff := oldLen - newLen
	if diff < 0 { // increase
		m.start = max(m.start+diff, 0)
	} else { // decrease
		first := m.cursor - m.start
		m.start = m.start + clamp(diff, 0, first)
	}
	m.end = min(m.start+newLen, len(m.VisualRows()))
}

// UpdateContent updates the list content based on the previously defined
// columns and rows.
// OPTIM: update-content cannot reflect on previous state to determine what rows
// actually require new rendering; therefore, it renders everything, even if the
// row was already included in the viewport contents and its selection-state did
// not change.
func (m *Model) UpdateContent() (updateHeader bool) {
	if m.Height() < 1 || m.cursor < 0 {
		return
	}

	// Render only rows that fit within the viewport
	// Constant runtime, independent of number of rows in a table.
	// Limits the number of renderedRows to a maximum of m.viewport.Height

	// TODO: consider combining loops
	var colChanged bool
	if m.dynCols {
		for j := range m.cols {
			mx := len(m.cols[j].Title) + len(m.cols[j].Suffix)
			for i := m.start; i < m.end; i++ {
				mx = max(mx, len(m.VisualRows()[i].Fields[j].Value()))
			}
			colChanged = colChanged || mx != m.cols[j].Width // once true, stays true
			m.cols[j].DynamicWidth = mx
		}
	}

	renderedRows := make([]string, 0, max(0, m.end-m.start))
	for i := m.start; i < m.end; i++ {
		renderedRows = append(renderedRows, m.renderRow(i))
	}

	m.content.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)

	// ensures horizontal position is updated if content width changed
	m.content.ScrollLeft(0)

	return colChanged
}

func (m *Model) UpdateHeader() {
	headerRow := m.renderHeader()
	m.header.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, headerRow),
	)
	// ensures horizontal position is updated if header width changed
	m.header.ScrollLeft(0)
}

// Focus focuses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
	m.UpdateContent()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
	m.UpdateContent()
}

// SetRows sets a new rows state.Can be unsafe if the number of columns changes.
// Use SetContent if rows and columns change together.
func (m *Model) SetRows(r []Row) {
	m.rows = r

	// in case of row count reduction, ensure m.end & cursor are in sync with new set
	m.end = min(m.end, len(m.VisualRows()))
	m.start = max(0, m.end-m.Height())

	if m.cursor > len(m.VisualRows())-1 {
		m.SetCursor(len(m.VisualRows()) - 1)
	}

	m.updateContentHeight()
	m.UpdateContent()
}

// AppendRows appends rows to the table's state. This can be unsafe if the
// number of columns is not equal to that of existing rows.
func (m *Model) AppendRows(r []Row) {
	rr := make([]Row, len(m.rows)+len(r))
	copy(rr[:len(m.rows)], m.rows)
	copy(rr[len(m.rows):], r)
	m.rows = rr

	m.updateContentHeight()
	m.UpdateContent()
}

// SetVirtualRows sets the virtual rows
// Note that supplying nil or [] does not reset the view.
// To completely remove virtual rows (even if empty) from view, use the
// `ResetVirtualRows` method.
func (m *Model) SetVirtualRows(r []Row) {
	if r == nil {
		r = []Row{} // ensure this function cannot be abused to replace ResetVirtualRows
	}
	if m.virtualRows == nil { // virtual rows come into view
		m.lastCursor = m.cursor
	}

	m.virtualRows = r

	// OPTIM: perhaps not ideal to hard-set each time
	m.start = 0
	m.end = clamp(m.start+m.content.Height(), m.start, len(r))

	if m.cursor > max(0, len(m.VisualRows())-1) {
		m.SetCursor(len(m.VisualRows()) - 1)
	}

	if colChanged := m.UpdateContent(); colChanged {
		m.UpdateHeader()
	}
}

// resetVirtaulRows empties virtual rows and ensures that the base rows are returned
func (m *Model) ResetVirtualRows() {
	m.virtualRows = nil
	m.cursor = m.lastCursor

	// ensure that new cursor location is visible
	if m.cursorOutOfBounds() {
		m.start = clamp(m.cursor, 0, max(0, len(m.rows)-m.content.Height()))
	}

	// ensures m.start & m.end are readjusted to newly visible rows
	m.MoveContentBoundaries(0)

	if colChanged := m.UpdateContent(); colChanged {
		m.UpdateHeader()
	}
}

// SetContent is suited when columns and rows change simultaneously, especially
// when the number of columns changes from the previous content state. This
// operation also completely resets any virtual rows.
func (m *Model) SetContent(c []Column, r []Row) {
	m.ResetVirtualRows()

	m.cols = c
	m.rows = r

	if m.cursor > len(m.VisualRows())-1 {
		m.SetCursor(len(m.VisualRows()) - 1)
	}
	m.updateContentHeight()
	m.UpdateContent()
	m.UpdateHeader()
}

// SetColumns sets a new columns state. Can be unsafe if the number of columns
// changes. Use SetContent if rows and columns change together.
func (m *Model) SetColumns(c []Column) {
	m.cols = c
	m.UpdateContent()
	m.UpdateHeader()
}

// SetFieldDelegate sets the field-delegate function
func (m *Model) SetFieldDelegate(f FieldDelegate) {
	m.fieldDelegate = f
}

// SetHeaderDelegate sets the header-delegate function
func (m *Model) SetHeaderDelegate(f HeaderDelegate) {
	m.headerDelegate = f
}

// SetDynamicColumnWidth updates the setting for dynamic-column-width and
// updates the view appropriately
func (m *Model) SetDynamicColumnWidth(b bool) {
	m.dynCols = b
	m.UpdateContent()
	m.UpdateHeader()
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.content.SetWidth(w)
	m.header.SetWidth(w)
	m.UpdateContent()
	m.UpdateHeader()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.content.SetHeight(h - m.header.Height())
	m.updateContentHeight()
	if colChanged := m.UpdateContent(); colChanged {
		m.UpdateHeader()
	}
}

// SetCursor sets the cursor position in the table.
func (m *Model) SetCursor(n int) {
	n = clamp(n, 0, max(0, len(m.VisualRows())-1))
	if m.cursor < n {
		m.MoveUp(n - m.cursor)
	} else if m.cursor > n {
		m.MoveDown(m.cursor - n)
	}
}

// MoveUp moves the selection up by any number of rows.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, max(0, len(m.VisualRows())-1))
	if m.cursorOutOfBounds() {
		m.MoveContentBoundaries(-n)
	}
	if colChanged := m.UpdateContent(); colChanged {
		m.UpdateHeader()
	}
}

// MoveDown moves the selection down by any number of rows.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, max(0, len(m.VisualRows())-1))
	if m.cursorOutOfBounds() {
		m.MoveContentBoundaries(n)
	}
	if colChanged := m.UpdateContent(); colChanged {
		m.UpdateHeader()
	}
}

func (m *Model) cursorOutOfBounds() bool {
	return m.cursor < m.start || m.cursor >= m.end
}

// ScrollRight scrolls the header and viewport contents to the right
func (m *Model) ScrollRight(n int) {
	m.header.ScrollRight(n)
	m.content.ScrollRight(n)
}

// ScrollLeft scrolls the header and viewport contents to the left
func (m *Model) ScrollLeft(n int) {
	m.header.ScrollLeft(n)
	m.content.ScrollLeft(n)
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.VisualRows()) - 1)
}

// GotoLeft scrolls back to the row beginning
func (m *Model) GotoLeft() {
	m.header.SetXOffset(0)
	m.content.SetXOffset(0)
}

// GotoRight scrolls to the rows ending
func (m *Model) GotoRight() {
	m.header.SetXOffset(math.MaxInt)
	m.content.SetXOffset(math.MaxInt)
}

// FromValues create the table rows from a simple string. It uses `\n` by
// default for getting all the rows and the given separator for the fields on
// each row.
// This does not apply styling
func (m *Model) FromValues(value, separator string, cb func(v string) Field) {
	rows := []Row{} //nolint:prealloc
	for _, line := range strings.Split(value, "\n") {
		r := Row{}
		for _, fieldValue := range strings.Split(line, separator) {
			r.Fields = append(r.Fields, cb(fieldValue))
		}
		rows = append(rows, r)
	}

	m.SetRows(rows)
}

func (m *Model) renderHeader() string {
	s := make([]string, 0, len(m.cols))
	for i, col := range m.cols {
		if col.InVisible {
			continue
		}
		width := ternary(col.DynamicWidth, col.Width, m.dynCols && col.DynamicWidth > 0)
		if width <= 0 {
			continue
		}

		// apply header-delegate if available
		if m.headerDelegate != nil {
			s = append(s, m.headerDelegate(col, i, width, m.styles.Header.GetPaddingLeft(), m.styles.Header.GetPaddingRight()))
			continue
		}

		// proceed with default styling if not
		style := lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true)
		renderedCell := style.Render(fmt.Sprintf(
			"%s%s",
			ansi.Truncate(col.Title, width-len(col.Suffix), "…"),
			col.Suffix,
		))
		s = append(s, m.styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, slices.Clip(s)...)
}

func (m *Model) renderRow(r int) string {
	s := make([]string, 0, len(m.cols))
	rows := m.VisualRows()
	for i := range rows[r].Fields {
		if m.cols[i].InVisible {
			continue
		}
		width := ternary(m.cols[i].DynamicWidth, m.cols[i].Width, m.dynCols && m.cols[i].DynamicWidth > 0)
		if width <= 0 {
			continue
		}

		// apply field-delegate if available
		if m.fieldDelegate != nil {
			s = append(s, m.fieldDelegate(rows[r], m.cols[i], i, r, width, m.styles.Cell.GetPaddingLeft(), m.styles.Cell.GetPaddingRight(), r == m.cursor))
			continue
		}

		// proceed with default styling if not

		value := rows[r].Fields[i].Value()
		enforceWidth := lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true).Render
		renderedCell := m.styles.Cell.Render(enforceWidth(ternary(value, ansi.Truncate(value, width, "…"), m.dynCols)))

		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, slices.Clip(s)...)

	if r == m.cursor {
		return m.styles.Selected.Render(row)
	}

	return row
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}

func ternary[T any](first T, second T, cond bool) T {
	if cond {
		return first
	}
	return second
}
