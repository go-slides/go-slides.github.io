package playground

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-slides/go-slides.github.io/playground/store"
	"github.com/go-widgets/mvvm"
)

// The file panel: the project's files beside the deck — slides.yaml, the
// themes, the kinds, data files and media — in the draft project of the
// store (draft.go), so they are kept across reloads with the deck. The
// layout is the one the slides command and the VS Code extension read:
// unzipped, the project builds with `slides build deck.md`.

// DeckFile is the deck's path in a project.
const DeckFile = store.DeckPath

// ProjectDirs are the folders the panel always shows.
var ProjectDirs = []string{"kinds", "media", "themes"}

// projectRoot is the folder the renderer sees the project in: the deck is
// DeckPath, the rest beside it.
var projectRoot = path.Dir(DeckPath)

// MaxTextFiles bounds the text the renderer is sent besides the deck, on
// every keystroke.
const MaxTextFiles = 2 << 20

// ProjectFile is one entry of the file panel.
type ProjectFile struct {
	Path string
	Dir  bool
	Size int64
}

// fileState is the file panel's part of the view-model.
type fileState struct {
	// Files are the panel's entries: the deck, the project's files and its
	// folders.
	Files *mvvm.Observable[[]ProjectFile]
	// Open is the file the editor shows: DeckFile, or a text file of the
	// project, whose text is FileText.
	Open     *mvvm.Observable[string]
	FileText *mvvm.Observable[string]
	// Usage says how much of the project's bounds the files take.
	Usage *mvvm.Observable[string]
	// MediaVersion counts the changes of the media, which the page posts
	// into the preview (Media).
	MediaVersion *mvvm.Observable[int]
	// ProjectAction is the Project menu's selection (ProjectItems): 0 its
	// title, another index an action, run when chosen. CreateAsked asks the
	// panel to start a new name.
	ProjectAction *mvvm.Observable[int]
	CreateAsked   *mvvm.Observable[CreateAsk]

	// FilesSaved, set by the host, is told when the files changed: it saves
	// the draft a moment later, as it does after typing.
	FilesSaved func()

	filesOf     *store.Project  // the project the panel shows
	dirs        map[string]bool // folders made in the panel, kept while empty
	mediaSig    string
	fileSyncing bool
}

// initFiles makes the panel's observables; NewViewModel calls it.
func (vm *ViewModel) initFiles() {
	vm.Files = mvvm.NewObservableEq[[]ProjectFile](nil, slices.Equal[[]ProjectFile])
	vm.Open = mvvm.NewObservable(DeckFile)
	vm.FileText = mvvm.NewObservable("")
	vm.Usage = mvvm.NewObservable("")
	vm.MediaVersion = mvvm.NewObservable(0)
	vm.ProjectAction = mvvm.NewObservable(0)
	vm.CreateAsked = mvvm.NewObservable(CreateAsk{})
	vm.ProjectAction.Subscribe(vm.chooseProjectAction)
	vm.FileText.Subscribe(func(s string) {
		p := vm.Open.Get()
		if vm.fileSyncing || p == DeckFile {
			return
		}
		if err := vm.project().Write(p, []byte(s), vm.drafts.now()); err != nil {
			vm.notice("⚠ " + err.Error())
			return
		}
		vm.filesChanged()
	})
	vm.followProject()
}

// ProjectThemeChoices are ThemeChoices with the themes the project's style
// sheets define (/* @theme name */), after the built-in ones.
func ProjectThemeChoices(src, css string) []string {
	out := ThemeChoices(src)
	for _, m := range atTheme.FindAllStringSubmatch(css, 32) {
		if !slices.Contains(out, m[1]) {
			out = slices.Insert(out, len(BuiltinThemes), m[1])
		}
	}
	return out
}

// followProject takes the files of a draft that replaced the one shown (a
// draft restored, the drafts forgotten): the folders it keeps, and the deck
// in the editor when the open file is not in it.
func (vm *ViewModel) followProject() {
	p := vm.project()
	if p == vm.filesOf {
		return
	}
	vm.filesOf = p
	vm.dirs = map[string]bool{}
	for d := range strings.SplitSeq(p.UI["dirs"], "\n") {
		if c, err := store.CleanPath(d); err == nil {
			vm.dirs[c] = true
		}
	}
	if o := vm.Open.Get(); o != DeckFile {
		if _, ok := p.File(o); !ok {
			vm.Open.Set(DeckFile)
		}
	}
	vm.refreshFiles()
}

// project is the draft project the files are kept in (made when the page
// keeps none).
func (vm *ViewModel) project() *store.Project {
	if vm.drafts.mine == nil {
		vm.drafts.mine = store.NewProject(DraftProject)
	}
	return vm.drafts.mine
}

// filesChanged is a change to the project's files: the panel, the theme
// choices and the preview follow, and the draft is saved with them.
func (vm *ViewModel) filesChanged() {
	vm.drafts.dirty = true
	vm.refreshFiles()
	vm.syncTheme()
	vm.changed()
	if vm.FilesSaved != nil {
		vm.FilesSaved()
	}
}

// refreshFiles lists the files for the panel and says the usage.
func (vm *ViewModel) refreshFiles() {
	p := vm.project()
	out := []ProjectFile{{Path: DeckFile, Size: int64(len(vm.Source.Get()))}}
	var sig strings.Builder
	for _, k := range p.Paths() {
		if k == DeckFile {
			continue
		}
		f, _ := p.File(k)
		out = append(out, ProjectFile{Path: k, Size: int64(len(f.Data))})
		if IsMedia(k) {
			fmt.Fprintf(&sig, "%s %d %d\n", k, len(f.Data), f.Modified.UnixNano())
		}
	}
	folders := vm.folders()
	for _, d := range folders {
		out = append(out, ProjectFile{Path: d, Dir: true})
	}
	vm.Files.Set(out)
	n := len(out) - len(folders)
	size := p.Size() - len(p.Text(DeckFile)) + len(vm.Source.Get())
	vm.Usage.Set(fmt.Sprintf("%d file%s · %s of %d MiB", n, plural(n), sizeText(size), store.MaxProjectSize>>20))
	if s := sig.String(); s != vm.mediaSig {
		vm.mediaSig = s
		vm.MediaVersion.Set(vm.MediaVersion.Get() + 1)
	}
}

// folders are the panel's folders: the usual three and those made in it.
func (vm *ViewModel) folders() []string {
	out := slices.Clone(ProjectDirs)
	for d := range vm.dirs {
		if !slices.Contains(out, d) {
			out = append(out, d)
		}
	}
	slices.Sort(out)
	return out
}

// saveDirs keeps the folders made in the panel with the project.
func (vm *ViewModel) saveDirs() {
	var ds []string
	for d := range vm.dirs {
		ds = append(ds, d)
	}
	slices.Sort(ds)
	vm.project().SetUI("dirs", strings.Join(ds, "\n"))
}

// Media are the project's media, path → bytes, for the preview.
func (vm *ViewModel) Media() map[string][]byte {
	p := vm.project()
	out := map[string][]byte{}
	for _, k := range p.Paths() {
		if IsMedia(k) {
			out[k], _ = p.Read(k)
		}
	}
	return out
}

// Modified is when a project file was last written, in Unix nanoseconds (0
// for none): the page keeps its copy of a medium while it holds.
func (vm *ViewModel) Modified(p string) int64 {
	if f, ok := vm.project().File(p); ok {
		return f.Modified.UnixNano()
	}
	return 0
}

// textFiles are the project's text files the renderer reads, by their path
// beside the deck, and the folders, within MaxTextFiles.
func (vm *ViewModel) textFiles() (map[string]string, []string) {
	p := vm.project()
	files := map[string]string{}
	total := 0
	for _, k := range p.Paths() {
		if k == DeckFile || IsMedia(k) {
			continue
		}
		b, _ := p.Read(k)
		if total += len(b); total > MaxTextFiles || !utf8.Valid(b) {
			continue
		}
		files[path.Join(projectRoot, k)] = string(b)
	}
	var dirs []string
	for _, d := range vm.folders() {
		dirs = append(dirs, path.Join(projectRoot, d))
	}
	return files, dirs
}

// hasFiles reports whether the project holds more than the deck.
func (vm *ViewModel) hasFiles() bool {
	for _, k := range vm.project().Paths() {
		if k != DeckFile {
			return true
		}
	}
	return false
}

// projectCSS is the project's style sheets, for the theme choices.
func (vm *ViewModel) projectCSS() string {
	var b strings.Builder
	p := vm.project()
	for _, k := range p.Paths() {
		if strings.EqualFold(path.Ext(k), ".css") {
			b.WriteString(p.Text(k))
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// isFolder reports whether p is one of the panel's folders, or holds files.
func (vm *ViewModel) isFolder(p string) bool {
	if slices.Contains(vm.folders(), p) {
		return true
	}
	for _, k := range vm.project().Paths() {
		if strings.HasPrefix(k, p+"/") {
			return true
		}
	}
	return false
}

// exists reports whether p names a file or a folder of the project.
func (vm *ViewModel) exists(p string) bool {
	if p == DeckFile {
		return true
	}
	_, ok := vm.project().File(p)
	return ok || vm.isFolder(p)
}

// OpenFile shows a file in the editor: the deck, or a text file; a medium
// is not text, and the status says how the deck names it.
func (vm *ViewModel) OpenFile(p string) {
	switch {
	case p == DeckFile:
		vm.Open.Set(DeckFile)
	case IsMedia(p):
		vm.notice(p + " is a medium: the deck shows it with " + MediaMarkdown(p) + " (drop it on the editor, or its menu: Insert into the deck).")
	case vm.isFolder(p):
	default:
		vm.fileSyncing = true
		vm.FileText.Set(vm.project().Text(p))
		vm.fileSyncing = false
		vm.Open.Set(p)
	}
}

// fail says an error, and reports whether there was one.
func (vm *ViewModel) fail(err error) bool {
	if err != nil {
		vm.notice("⚠ " + err.Error())
		return true
	}
	return false
}

// CreateFile makes a file with the given text and opens it; a theme, a kinds
// or a data file is listed in slides.yaml, where the slides command finds it.
func (vm *ViewModel) CreateFile(p, text string) error {
	c, err := store.CleanPath(p)
	if err != nil {
		return err
	}
	if vm.exists(c) {
		return fmt.Errorf("%s exists already", c)
	}
	if err := vm.project().Write(c, []byte(text), vm.drafts.now()); err != nil {
		return err
	}
	vm.listInConfig(c)
	vm.filesChanged()
	vm.OpenFile(c)
	return nil
}

// CreateFolder makes an empty folder.
func (vm *ViewModel) CreateFolder(p string) error {
	c, err := store.CleanPath(p)
	if err != nil {
		return err
	}
	if vm.exists(c) {
		return fmt.Errorf("%s exists already", c)
	}
	vm.dirs[c] = true
	vm.saveDirs()
	vm.filesChanged()
	return nil
}

// errDeck is a change the deck refuses: it is deck.md, by which name the
// slides command, the extension and this page read it.
var errDeck = errors.New("the deck stays deck.md: the slides command and the playground read it by that name")

// RenamePath renames or moves a file or a folder (everything in it), and the
// paths slides.yaml lists with it. The deck's text is not changed: a medium
// moved is named by its new path there by hand.
func (vm *ViewModel) RenamePath(from, to string) error {
	if from == DeckFile || to == DeckFile {
		return errDeck
	}
	t, err := store.CleanPath(to)
	if err != nil {
		return err
	}
	if vm.exists(t) {
		return fmt.Errorf("%s exists already", t)
	}
	p, now := vm.project(), vm.drafts.now()
	moved := 0
	for _, k := range p.Paths() {
		if k != from && !strings.HasPrefix(k, from+"/") {
			continue
		}
		b, _ := p.Read(k)
		if err := p.Write(t+k[len(from):], b, now); err != nil {
			return err
		}
		p.Remove(k)
		moved++
	}
	for d := range vm.dirs {
		if d == from || strings.HasPrefix(d, from+"/") {
			delete(vm.dirs, d)
			vm.dirs[t+d[len(from):]] = true
			moved++
		}
	}
	builtin := slices.Contains(ProjectDirs, from)
	if moved == 0 && !builtin {
		return fmt.Errorf("%s: no such file", from)
	}
	if builtin {
		vm.dirs[t] = true
	}
	vm.saveDirs()
	vm.renameInConfig(from, t, false)
	if o := vm.Open.Get(); o == from || strings.HasPrefix(o, from+"/") {
		vm.Open.Set(t + o[len(from):])
	}
	vm.filesChanged()
	return nil
}

// DeletePath deletes a file, or a folder and what it holds, and drops them
// from slides.yaml.
func (vm *ViewModel) DeletePath(target string) error {
	if target == DeckFile {
		return errDeck
	}
	p := vm.project()
	for _, k := range p.Paths() {
		if k == target || strings.HasPrefix(k, target+"/") {
			p.Remove(k)
		}
	}
	for d := range vm.dirs {
		if d == target || strings.HasPrefix(d, target+"/") {
			delete(vm.dirs, d)
		}
	}
	vm.saveDirs()
	vm.renameInConfig(target, "", true)
	if o := vm.Open.Get(); o == target || strings.HasPrefix(o, target+"/") {
		vm.Open.Set(DeckFile)
	}
	vm.filesChanged()
	return nil
}

// AddFile stores bytes taken from outside (a drop, a file picked) in folder
// dir under name; a medium dropped at the top goes in media/. A name taken
// by other bytes gets a number; the same bytes again are the same file. It
// returns the path.
func (vm *ViewModel) AddFile(dir, name string, data []byte) (string, error) {
	base := path.Base(strings.ReplaceAll(name, `\`, "/"))
	if dir == "" && IsMedia(base) {
		dir = "media"
	}
	c, err := store.CleanPath(path.Join(dir, base))
	if err != nil || base == "." || base == "/" {
		return "", fmt.Errorf("%q is not a file name", name)
	}
	p := vm.project()
	ext := path.Ext(c)
	stem := strings.TrimSuffix(c, ext)
	for i := 2; ; i++ {
		old, ok := p.Read(c)
		if !ok && !vm.isFolder(c) && c != DeckFile {
			break
		}
		if ok && string(old) == string(data) {
			return c, nil
		}
		c = stem + "-" + strconv.Itoa(i) + ext
	}
	if err := p.Write(c, data, vm.drafts.now()); err != nil {
		return "", err
	}
	vm.listInConfig(c)
	vm.filesChanged()
	return c, nil
}

// uniqueName is the first of name, name-2, name-3… whose file is free.
func (vm *ViewModel) uniqueName(dir, name, ext string) (string, string) {
	n := name
	for i := 2; vm.exists(path.Join(dir, n+ext)); i++ {
		n = name + "-" + strconv.Itoa(i)
	}
	return n, path.Join(dir, n+ext)
}

// NewTheme makes a theme built on base in themes/ and opens it.
func (vm *ViewModel) NewTheme(base string) {
	if !slices.Contains(BuiltinThemes, base) {
		base = "default"
	}
	name, p := vm.uniqueName("themes", "my-"+base, ".css")
	if vm.fail(vm.CreateFile(p, ThemeTemplate(name, base))) {
		return
	}
	vm.notice("Made " + p + ", built on " + base + ": choose " + name + " in the Theme list to use it.")
}

// NewKind makes a kinds file in kinds/ and opens it.
func (vm *ViewModel) NewKind() {
	name, p := vm.uniqueName("kinds", "chapter", ".yaml")
	if vm.fail(vm.CreateFile(p, KindTemplate(name))) {
		return
	}
	vm.notice("Made " + p + ": a slide takes the kind with <!-- _kind: " + name + " -->.")
}

// NewDataFile makes a template data file in data/ and opens it.
func (vm *ViewModel) NewDataFile() {
	_, p := vm.uniqueName("data", "values", ".yaml")
	if vm.fail(vm.CreateFile(p, DataTemplate())) {
		return
	}
	vm.notice("Made " + p + ": the deck reads it as {{ .title }}.")
}

// listInConfig lists a new theme, kinds or data file in slides.yaml, unless
// the project's configuration is slides.hcl (edited by hand).
func (vm *ViewModel) listInConfig(p string) {
	key, value := "", ""
	ext := strings.ToLower(path.Ext(p))
	switch {
	case strings.HasPrefix(p, "themes/") && ext == ".css":
		key, value = "themes", "themes"
	case strings.HasPrefix(p, "kinds/") && (ext == ".yaml" || ext == ".yml" || ext == ".hcl"):
		key, value = "kinds", p
	case strings.HasPrefix(p, "data/") && (ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".hcl"):
		key, value = "data", p
	default:
		return
	}
	vm.editConfig(func(src string) (string, error) { return ListInConfig(src, key, value) })
}

// renameInConfig follows a rename or a deletion in slides.yaml.
func (vm *ViewModel) renameInConfig(from, to string, remove bool) {
	vm.editConfig(func(src string) (string, error) { return RenameInConfig(src, from, to, remove) })
}

// editConfig rewrites slides.yaml through edit; the editor follows when it
// shows the file.
func (vm *ViewModel) editConfig(edit func(string) (string, error)) {
	p := vm.project()
	if _, hcl := p.File("slides.hcl"); hcl {
		return
	}
	old := p.Text(ConfigPath)
	src, err := edit(old)
	if vm.fail(err) || src == old {
		return
	}
	if vm.fail(p.Write(ConfigPath, []byte(src), vm.drafts.now())) {
		return
	}
	if vm.Open.Get() == ConfigPath {
		vm.fileSyncing = true
		vm.FileText.Set(src)
		vm.fileSyncing = false
	}
}

// projectFiles are the project as it goes out: the deck as it is typed,
// every file, the empty folders.
func (vm *ViewModel) projectFiles() ProjectFiles {
	p := vm.project()
	out := ProjectFiles{Files: map[string][]byte{DeckFile: []byte(vm.Source.Get())}}
	for _, k := range p.Paths() {
		if k != DeckFile {
			out.Files[k], _ = p.Read(k)
		}
	}
	for _, d := range vm.folders() {
		held := false
		for k := range out.Files {
			if strings.HasPrefix(k, d+"/") {
				held = true
				break
			}
		}
		if !held {
			out.Dirs = append(out.Dirs, d)
		}
	}
	return out
}

// ZipName is the name a downloaded project has.
const ZipName = "slides-project.zip"

// ExportZip is the project as a .zip.
func (vm *ViewModel) ExportZip() ([]byte, error) {
	return ExportZip(vm.projectFiles(), vm.modified)
}

// downloadZip offers the project as a .zip.
func (vm *ViewModel) downloadZip() {
	z, err := vm.ExportZip()
	if vm.fail(err) {
		return
	}
	d, ok := vm.host.(BytesDownloader)
	if !ok {
		vm.notice("⚠ this page cannot save a file")
		return
	}
	d.DownloadBytes(ZipName, "application/zip", z)
	vm.notice("Saved " + ZipName + ": unzipped, `slides build deck.md` builds it, and the VS Code extension opens it.")
}

// modified is when a file was last written (now for the deck).
func (vm *ViewModel) modified(p string) time.Time {
	if f, ok := vm.project().File(p); ok && p != DeckFile {
		return f.Modified
	}
	return vm.drafts.now()
}

// Picker is a host that lets the reader pick files: what is "media" (files
// to add to media/), "zip" (a project's .zip) or "folder" (a project's
// folder). The host hands what was picked to AddFile, OpenZip or OpenFolder.
type Picker interface {
	Pick(what string)
}

// BytesDownloader is a host that saves binary files (a .zip).
type BytesDownloader interface {
	DownloadBytes(name, mime string, data []byte)
}

// CreateAsk asks the file panel to start typing a new name in place (the
// Project menu's New file… and New folder…): N counts the asks.
type CreateAsk struct {
	Folder bool
	N      int
}

// ProjectItems are the Project menu's items; item 0 is its title.
var ProjectItems = []string{
	"Project",
	"New theme (from default)", "New theme (from gaia)", "New theme (from uncover)",
	"New kind", "New data file", "New file…", "New folder…",
	"Add media…", "Open project (.zip)…", "Open project folder…", "Download project (.zip)",
}

// chooseProjectAction runs a Project menu item and returns the menu to its
// title.
func (vm *ViewModel) chooseProjectAction(i int) {
	if i <= 0 || i >= len(ProjectItems) {
		return
	}
	vm.ProjectAction.Set(0)
	pick := func(what string) {
		if p, ok := vm.host.(Picker); ok {
			p.Pick(what)
		}
	}
	switch i {
	case 1, 2, 3:
		vm.NewTheme(BuiltinThemes[i-1])
	case 4:
		vm.NewKind()
	case 5:
		vm.NewDataFile()
	case 6, 7:
		vm.CreateAsked.Set(CreateAsk{Folder: i == 7, N: vm.CreateAsked.Get().N + 1})
	case 8:
		pick("media")
	case 9:
		pick("zip")
	case 10:
		pick("folder")
	case 11:
		vm.downloadZip()
	}
}

// CreateFromPanel makes a file typed in the panel: a theme, kind or data
// file made in its folder starts from its template.
func (vm *ViewModel) CreateFromPanel(p string) error {
	name := strings.TrimSuffix(path.Base(p), path.Ext(p))
	text := ""
	switch ext := strings.ToLower(path.Ext(p)); {
	case strings.HasPrefix(p, "themes/") && ext == ".css" && themeName.MatchString(name):
		text = ThemeTemplate(name, "default")
	case strings.HasPrefix(p, "kinds/") && (ext == ".yaml" || ext == ".yml") && themeName.MatchString(name):
		text = KindTemplate(name)
	case strings.HasPrefix(p, "data/") && (ext == ".yaml" || ext == ".yml"):
		text = DataTemplate()
	}
	return vm.CreateFile(p, text)
}

// DownloadFile offers one file of the project.
func (vm *ViewModel) DownloadFile(p string) {
	b, ok := vm.project().Read(p)
	if p == DeckFile {
		b, ok = []byte(vm.Source.Get()), true
	}
	d, can := vm.host.(BytesDownloader)
	if !ok || !can {
		return
	}
	mime := "text/plain"
	if IsMedia(p) {
		mime = MediaType(p, b)
	}
	d.DownloadBytes(path.Base(p), mime, b)
}

// OpenZip replaces the project with a .zip's, after asking; the draft it
// replaces stays available (Draft ▸ Restore replaced).
func (vm *ViewModel) OpenZip(data []byte) {
	pf, err := ImportZip(data)
	if vm.fail(err) {
		return
	}
	vm.openProject(pf, "the .zip")
}

// OpenFolder replaces the project with a folder's files.
func (vm *ViewModel) OpenFolder(files map[string][]byte) {
	pf, err := ImportFolder(files)
	if vm.fail(err) {
		return
	}
	vm.openProject(pf, "the folder")
}

// openProject puts a project in place of the draft.
func (vm *ViewModel) openProject(pf ProjectFiles, from string) {
	if c, ok := vm.host.(Confirmer); ok && !c.Confirm("Replace your deck and its files with "+from+"'s? What you have now stays in this browser: Draft ▸ Restore replaced.") {
		return
	}
	deck := string(pf.Files[DeckFile])
	if len(deck) > MaxSource {
		vm.fail(fmt.Errorf("the deck is %d KiB; the playground renders decks up to %d KiB", len(deck)>>10, MaxSource>>10))
		return
	}
	vm.archiveDraft(deck)
	p := store.NewProject(DraftProject)
	now := vm.drafts.now()
	for k, b := range pf.Files {
		if vm.fail(p.Write(k, b, now)) {
			return
		}
	}
	vm.drafts.mine = p
	vm.dirs = map[string]bool{}
	for _, d := range pf.Dirs {
		vm.dirs[d] = true
	}
	vm.saveDirs()
	vm.Open.Set(DeckFile)
	vm.Source.Set(deck)
	vm.filesChanged()
	vm.notice(fmt.Sprintf("Opened %s: %d files.", from, len(pf.Files)))
}
