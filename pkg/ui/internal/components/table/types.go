package table

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

// Model defines a state for the table widget.
type Model struct {
	KeyMap *KeyMap
	Help   help.Model

	// rowCache caches rendered-rows; it is reset upon any table update &
	// sideways scrolling
	rowCache map[int]string

	// fieldDelegate, when set, is called to provide non-default styling for a row
	fieldDelegate FieldDelegate
	// headerDelegate, when set, is called to provide non-default styling for the header
	headerDelegate HeaderDelegate

	cols []Column
	// TODO: if rows are already stored in memory in their entirety, can't they
	// 'simply' be passed off to the viewport in their entirety, instead of the
	// complicated viewport content updates that also re-renders every visible
	// line each time.
	rows []Row

	// virtual rows replace rows (visually) when len > 0
	// virtual rows can be used to show filtered data
	virtualRows []Row
	cursor      int
	lastCursor  int // used to reset cursor after a virtual row reset
	focus       bool
	styles      Styles

	content viewport.Model
	header  viewport.Model

	// start & end represent the row indices of the content that is offered to
	// the content viewport, which in this implementation is equal to the rows
	// that are visible.
	start int // inclusive
	end   int // exclusive
}

// FieldDelegate delegates the rendering of the row cell (including padding) to
// the implementer.
//
// It provides the following information:
//
// [row (Row)]
// The `Row` object containing the field being rendered.
//
// [col (Column)]
// The `Column` object containing column information, such as title and
// the visibility flag.
//
// [colIdx (int)]
// The column-index for the field currently being rendered. Can be used to
// access the field with `row.Fields[colIdx]`.
//
// [rowIdx (int)]
// The row-index of the row for which a field is currently being rendered.
//
// [colWidth (int)]
// The allowed width for cell contents, excluding padding.
//
// [padL (int)]
// The padding to be added to the left side of the rendered cell.
//
// [padR (int)]
// The padding to be added to the right side of the rendered cell.
//
// [selected (bool)]
// A boolean flag that indicates whether or not the field is part of the
// currently selected (a.k.a. focused) row.
//
// [inview (bool)]
// A boolean flag that indicates whether or not the field is (either in part or
// in full) within the boundaries of the viewport, i.e. in view for the user.
// Note that when rendering cells left of the viewport that are out-of-view, the
// width still needs to be respected to prevent unintended shifts when scrolling
// sideways through the data.
//
// [offL (int)]
// The number of terminal cells (not table cell, but the monospace cell for
// rendering text in the terminal emulator) left of the viewport leftmost border
// that are out of view. Only returns values in the inclusive range [0,
// column-width] where column-width includes padding.
//
// [offR (int)]
// The number of terminal cells (not table cell, but the monospace cell for
// rendering text in the terminal emulator) right of the viewport rightmost
// border that are out of view. Only returns values in the inclusive range [0,
// column-width] where column-width includes padding.
type FieldDelegate func(row Row, col Column, colIdx, rowIdx, colWidth, padL, padR int, selected, inview bool, offL, offR int) string

// HeaderDelegate delegates the rendering of the header cell (including padding) to
// the implementer.
//
// It provides the following information:
//
// [col (Column)]
// The `Column` object for which the header cell is currently being rendered.
//
// [colIdx (int)]
// The column-index for the header cell currently being rendered.
//
// [colWidth (int)]
// The allowed width for header cell contents, excluding padding.
//
// [padL (int)]
// The padding to be added to the left side of the rendered cell.
//
// [padR (int)]
// The padding to be added to the right side of the rendered cell.
//
// [inview (bool)]
// A boolean flag that indicates whether or not the header cell is (either in
// part or in full) within the boundaries of the viewport, i.e. in view for the
// user. Note that when rendering columns left of the viewport that are
// out-of-view, the width still needs to be respected to prevent unintended
// shifts when scrolling sideways through the data.
type HeaderDelegate func(col Column, colIdx, colWidth, padL, padR int, inview bool) string

// Row represents one line in the table.
type Row struct {
	Fields   []Field
	Metadata map[string]any
}

type Field interface {
	Value() string
}

func (r Row) String() string {
	res := strings.Builder{}
	for i, f := range r.Fields {
		if i > 0 {
			res.WriteString(" ")
		}
		res.WriteString(f.Value())
	}
	return res.String()
}

type Rows []Row

// Convenience function for obtaining plain string representation of each row.
// Useful for searching.
func (r Rows) ToStrings() []string {
	rows := []Row(r)
	res := make([]string, len(rows))
	for i := range rows {
		res[i] = rows[i].String()
	}
	return res
}

// Column defines the table structure.
type Column struct {
	Title           string
	Suffix          string
	Width           int
	DynamicWidth    int
	UseDynamicWidth bool
	InVisible       bool
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Header lipgloss.Style
	// only affects default styling, remains unused when using a delegate
	Cell lipgloss.Style
	// only affects default styling, remains unused when using a delegate
	Selected lipgloss.Style
}
