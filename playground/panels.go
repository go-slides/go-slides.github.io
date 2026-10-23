package playground

import "github.com/go-widgets/mvvm"

// The panes around the editor can be switched off, and the ribbon collapsed, so
// that a narrow window gives the editor and the preview the room: at 1280 px the
// editor with its file panel and the help beside it is thirty columns wide. What a
// person chose is kept with the rest of the UI state (saveUI); until they choose,
// the window's width decides (DefaultPanels).

// The widths (CSS pixels) from which a pane is on by default: with the file
// panel (190 px), the thumbnails (168 px) and the help (280 px) around them, the
// editor still gets about 60 columns.
const (
	filesMinWidth = 1440
	helpMinWidth  = 1700
)

// panelState is the visibility of the panes and of the ribbon.
type panelState struct {
	// FilesPane is the file panel, HelpPane the help and documentation column,
	// RibbonCollapsed the ribbon reduced to its tabs.
	FilesPane, HelpPane, RibbonCollapsed *mvvm.Observable[bool]
	// saved are the keys the stored UI state had: those the person chose.
	saved map[string]bool
	// quiet stops saveUI while defaults are set: they are not choices.
	quiet bool
}

func newPanelState() panelState {
	return panelState{
		FilesPane:       mvvm.NewObservable(true),
		HelpPane:        mvvm.NewObservable(true),
		RibbonCollapsed: mvvm.NewObservable(false),
		saved:           map[string]bool{},
	}
}

// DefaultPanels turns the file panel and the help on or off for a window cssWidth
// CSS pixels wide, for each the person has not chosen yet; the defaults are not
// stored, so a wider window later gets its panes back.
func (vm *ViewModel) DefaultPanels(cssWidth int) {
	vm.quiet = true
	defer func() { vm.quiet = false }()
	if !vm.saved["files"] {
		vm.FilesPane.Set(cssWidth >= filesMinWidth)
	}
	if !vm.saved["help"] {
		vm.HelpPane.Set(cssWidth >= helpMinWidth)
	}
}
