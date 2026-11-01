package playground

import (
	"strconv"

	"github.com/go-widgets/mvvm"
)

// The panes around the editor can be switched off, and the ribbon collapsed, so
// that a narrow window gives the editor and the preview the room: at 1280 px the
// editor with its file panel and the help beside it is thirty columns wide. What a
// person chose is kept with the rest of the UI state (saveUI); until they choose,
// the window's width decides (DefaultPanels).
//
// The panes are also resizable: the dividers between them (a toolkit.Splitter in
// the scene) write the widths below, and the editor:preview share. A size the
// person dragged, typed with the keys or collapsed is kept like a switch; the
// default widths are not stored, so they follow the defaults of a later version.

// The widths (CSS pixels) from which a pane is on by default: with the file
// panel (190 px), the thumbnails (168 px) and the help (280 px) around them, the
// editor still gets about 60 columns.
const (
	filesMinWidth = 1440
	helpMinWidth  = 1700
)

// The default width of each pane in logical (CSS) pixels, and the range the
// divider keeps it in; the editor and the preview share what is left, 6 to 5.
const (
	filesWidth, filesMin, filesMax    = 190, 120, 480
	thumbsWidth, thumbsMin, thumbsMax = 168, 96, 360
	helpWidth, helpMin, helpMax       = 280, 180, 560
	editorWeight, previewWeight       = 600, 500
	editorMin, previewMin             = 260, 220
)

// The panes' indexes in the splitter, and the keys of their sizes in the stored
// UI state.
const (
	paneFiles = iota
	paneEditor
	paneThumbs
	panePreview
	paneHelp
	paneVersions
	paneGit
)

// sizeKeys name the stored sizes: the three widths, and the editor's share of the
// room it divides with the preview (in thousandths).
const (
	keyFilesWidth  = "w-files"
	keyThumbsWidth = "w-thumbs"
	keyHelpWidth   = "w-help"
	// the Versions pane's (versions.go)
	keyVersionsWidth = "w-versions"
	// the GitHub pane's (scene_git.go)
	keyGitWidth = "w-git"
	keySplit    = "w-split"
)

// panelState is the visibility of the panes and of the ribbon, and their sizes.
type panelState struct {
	// FilesPane is the file panel, HelpPane the help and documentation column,
	// RibbonCollapsed the ribbon reduced to its tabs.
	FilesPane, HelpPane, RibbonCollapsed *mvvm.Observable[bool]
	// FilesWidth, ThumbsWidth and HelpWidth are the widths of the columns in
	// logical pixels (0: pulled shut by its divider); EditorWeight and
	// PreviewWeight what the editor and the preview share of the rest.
	FilesWidth, ThumbsWidth, HelpWidth *mvvm.Observable[int]
	// VersionsWidth is the Versions pane's width (scene_versions.go).
	VersionsWidth *mvvm.Observable[int]
	// GitWidth is the GitHub pane's width (scene_git.go).
	GitWidth                    *mvvm.Observable[int]
	EditorWeight, PreviewWeight *mvvm.Observable[int]
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
		FilesWidth:      mvvm.NewObservable(filesWidth),
		ThumbsWidth:     mvvm.NewObservable(thumbsWidth),
		HelpWidth:       mvvm.NewObservable(helpWidth),
		VersionsWidth:   mvvm.NewObservable(versionsWidth),
		GitWidth:        mvvm.NewObservable(gitWidth),
		EditorWeight:    mvvm.NewObservable(editorWeight),
		PreviewWeight:   mvvm.NewObservable(previewWeight),
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

// widthRange is a stored width: where it lives and what it may be.
type widthRange struct {
	key      string
	obs      *mvvm.Observable[int]
	min, max int
}

func (vm *ViewModel) widths() []widthRange {
	return []widthRange{
		{keyFilesWidth, vm.FilesWidth, filesMin, filesMax},
		{keyThumbsWidth, vm.ThumbsWidth, thumbsMin, thumbsMax},
		{keyHelpWidth, vm.HelpWidth, helpMin, helpMax},
		{keyVersionsWidth, vm.VersionsWidth, versionsMin, versionsMax},
		{keyGitWidth, vm.GitWidth, gitMin, gitMax},
	}
}

// EditorShare is the editor's part of the room it divides with the preview, in
// thousandths.
func (vm *ViewModel) EditorShare() int {
	e, p := vm.EditorWeight.Get(), vm.PreviewWeight.Get()
	if e+p <= 0 {
		return 1000 * editorWeight / (editorWeight + previewWeight)
	}
	return (e*1000 + (e+p)/2) / (e + p)
}

// PanesResized is the splitter telling that the person changed these sizes (a
// drag, a key, a double-click): keys are the stored keys of the panes it moved. They
// are kept from now on, the others stay defaults.
func (vm *ViewModel) PanesResized(keys []string) {
	for _, k := range keys {
		vm.saved[k] = true
	}
	vm.saveUI()
}

// saveSizes writes the sizes the person chose into the stored UI state p.
func (vm *ViewModel) saveSizes(set func(key, value string)) {
	for _, w := range vm.widths() {
		if vm.saved[w.key] {
			set(w.key, strconv.Itoa(w.obs.Get()))
		}
	}
	if vm.saved[keySplit] {
		set(keySplit, strconv.Itoa(vm.EditorShare()))
	}
}

// loadSizes reads the stored sizes back, each only if it is in range: a width
// between its Min and Max, or 0 (pulled shut); the share between a tenth and nine
// tenths.
func (vm *ViewModel) loadSizes(get func(key string) string) {
	vm.quiet = true
	defer func() { vm.quiet = false }()
	for _, w := range vm.widths() {
		if v, err := strconv.Atoi(get(w.key)); err == nil && (v == 0 || v >= w.min && v <= w.max) {
			w.obs.Set(v)
			vm.saved[w.key] = true
		}
	}
	if v, err := strconv.Atoi(get(keySplit)); err == nil && v >= 100 && v <= 900 {
		vm.EditorWeight.Set(v)
		vm.PreviewWeight.Set(1000 - v)
		vm.saved[keySplit] = true
	}
}
