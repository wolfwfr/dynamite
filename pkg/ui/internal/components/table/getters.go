package table

// Focused returns the focus state of the table.
func (m *Model) Focused() bool {
	return m.focus
}

// SelectedRow returns the selected row.
// You can cast it to your own implementation.
func (m *Model) SelectedRow() *Row {
	rows := m.VisualRows()
	if m.cursor < 0 || m.cursor >= len(rows) {
		return nil
	}

	return &rows[m.cursor]
}

// Visual rows returns virtual rows when set or falls back to rows
func (m *Model) VisualRows() []Row {
	if m.virtualRows != nil { // virtualRows with len 0 is valid
		return m.virtualRows
	}
	return m.rows
}

// Rows returns the current rows.
func (m *Model) Rows() []Row {
	return m.rows
}

// VirtualRows returns the current virtual rows.
func (m *Model) VirtualRows() []Row {
	return m.virtualRows
}

// Columns returns the current columns.
func (m *Model) Columns() []Column {
	return m.cols
}

// DynamicColumnWidth returns the current setting for dynamic-column-width
func (m *Model) DynamicColumnWidth() bool {
	return m.dynCols
}

// Height returns the viewport height of the table.
func (m *Model) Height() int {
	return m.content.Height() + m.header.Height()
}

// Width returns the viewport width of the table.
func (m *Model) Width() int {
	return m.content.Width()
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.cursor
}

// CursorAtEnd returns whether the selected row is the last available row
func (m *Model) CursorAtEnd() bool {
	return m.cursor == len(m.VisualRows())
}

// ViewAtEnd returns whether the last available row is in view
func (m *Model) ViewAtEnd() bool {
	return m.end == len(m.VisualRows())
}
