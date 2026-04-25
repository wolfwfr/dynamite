package messages

type View int

const (
	Table_selection View = iota
	Item_selection
)

type SwitchView struct {
	OldView View
	NewView View
}
