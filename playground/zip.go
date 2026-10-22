package playground

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/store"
)

// A project travels as a .zip whose layout is the one the slides command and
// the VS Code extension read as it is: deck.md at the top, slides.yaml next
// to it, themes/, kinds/, media/ and whatever else the deck names. Unzipped,
// `slides build deck.md` builds it.

// ProjectFiles are the files of a project going out or coming in: path →
// bytes, and the folders that hold nothing.
type ProjectFiles struct {
	Files map[string][]byte
	Dirs  []string
}

// stored are the media a .zip keeps as they are: they are compressed
// already, and deflating them again costs time for nothing.
var stored = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".avif": true,
	".mp4": true, ".webm": true, ".mov": true, ".m4v": true, ".ogg": true, ".mp3": true,
	".m4a": true, ".woff": true, ".woff2": true, ".zip": true,
}

// ExportZip writes a project as a .zip, files sorted, each with its time.
func ExportZip(p ProjectFiles, modified func(string) time.Time) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	names := make([]string, 0, len(p.Files))
	for k := range p.Files {
		names = append(names, k)
	}
	slices.Sort(names)
	for _, n := range names {
		method := zip.Deflate
		if stored[strings.ToLower(path.Ext(n))] {
			method = zip.Store
		}
		h := &zip.FileHeader{Name: n, Method: method, Modified: modified(n)}
		h.SetMode(0o644)
		f, err := w.CreateHeader(h)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write(p.Files[n]); err != nil {
			return nil, err
		}
	}
	dirs := slices.Clone(p.Dirs)
	slices.Sort(dirs)
	for _, d := range dirs {
		h := &zip.FileHeader{Name: d + "/", Method: zip.Store, Modified: modified(d)}
		h.SetMode(0o755 | 1<<31) // a directory
		if _, err := w.CreateHeader(h); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// The bounds of an imported .zip, past the project's own (store.MaxFiles,
// MaxFileSize, MaxProjectSize): how many entries the archive may list at
// all, how large it may be, and how much larger than its compressed size
// its contents may be (a zip bomb inflates a few kilobytes to gigabytes).
const (
	MaxZipEntries = 4 * store.MaxFiles
	MaxZipSize    = store.MaxProjectSize
	MaxZipRatio   = 100
	// ratioFloor is the size under which the ratio is not checked: a few
	// kilobytes of YAML deflate a lot, and are no danger.
	ratioFloor = 1 << 20
)

// ErrNoDeck is an archive with no deck in it.
var ErrNoDeck = errors.New("the archive holds no deck: put deck.md at its top (or one .md file)")

// ImportZip reads a project from a .zip, refusing what would leave the
// project (absolute paths, ..), what would exceed its bounds, and archives
// whose contents are out of proportion to their size. A single folder
// holding everything (a zipped project folder) is the project. Files are
// read through a limit, so a header that lies about its size gains nothing.
func ImportZip(data []byte) (ProjectFiles, error) {
	if len(data) > MaxZipSize {
		return ProjectFiles{}, fmt.Errorf("the archive is %d MiB: at most %d MiB", len(data)>>20, MaxZipSize>>20)
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ProjectFiles{}, fmt.Errorf("not a .zip archive: %v", err)
	}
	if len(r.File) > MaxZipEntries {
		return ProjectFiles{}, fmt.Errorf("the archive lists %d entries: at most %d", len(r.File), MaxZipEntries)
	}
	in := map[string][]byte{}
	var dirs []string
	var total, packed uint64
	for _, f := range r.File {
		name := f.Name
		if skipEntry(name) {
			continue
		}
		if f.FileInfo().IsDir() || strings.HasSuffix(name, "/") {
			c, err := store.CleanPath(strings.TrimSuffix(name, "/"))
			if err != nil {
				return ProjectFiles{}, err
			}
			dirs = append(dirs, c)
			continue
		}
		c, err := store.CleanPath(name)
		if err != nil {
			return ProjectFiles{}, err
		}
		if f.Mode()&(1<<27) != 0 { // a symbolic link
			return ProjectFiles{}, fmt.Errorf("%s: links are not taken", c)
		}
		if len(in) >= store.MaxFiles {
			return ProjectFiles{}, fmt.Errorf("a project holds at most %d files", store.MaxFiles)
		}
		if f.UncompressedSize64 > store.MaxFileSize {
			return ProjectFiles{}, fmt.Errorf("%s is over %d KiB: a file is at most that", c, store.MaxFileSize>>10)
		}
		b, err := readEntry(f)
		if err != nil {
			return ProjectFiles{}, fmt.Errorf("%s: %v", c, err)
		}
		total += uint64(len(b))
		packed += f.CompressedSize64
		if total > store.MaxProjectSize {
			return ProjectFiles{}, fmt.Errorf("the project would be over %d MiB", store.MaxProjectSize>>20)
		}
		if total > ratioFloor && total > MaxZipRatio*max(packed, 1) {
			return ProjectFiles{}, fmt.Errorf("the archive inflates more than %d times: refused as a possible zip bomb", MaxZipRatio)
		}
		if _, dup := in[c]; dup {
			return ProjectFiles{}, fmt.Errorf("%s is in the archive twice", c)
		}
		in[c] = b
	}
	return projectFrom(in, dirs)
}

// readEntry reads a file through a limit one byte past the largest file.
func readEntry(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(io.LimitReader(rc, store.MaxFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(b) > store.MaxFileSize {
		return nil, fmt.Errorf("more than %d KiB", store.MaxFileSize>>10)
	}
	return b, nil
}

// skipEntry is what an archiver adds that is not the project: macOS's
// resource forks and folder settings, Windows' thumbnails.
func skipEntry(name string) bool {
	base := path.Base(strings.TrimSuffix(name, "/"))
	return strings.HasPrefix(name, "__MACOSX/") || base == ".DS_Store" || base == "Thumbs.db" || strings.HasPrefix(base, "._")
}

// projectFrom makes a project of files read from an archive or a folder:
// the folder that holds everything is removed, and the deck is deck.md, or
// the only Markdown file at the top, renamed.
func projectFrom(in map[string][]byte, dirs []string) (ProjectFiles, error) {
	if top := commonTop(in, dirs); top != "" {
		out := make(map[string][]byte, len(in))
		for k, v := range in {
			out[strings.TrimPrefix(k, top+"/")] = v
		}
		in = out
		var ds []string
		for _, d := range dirs {
			if d != top {
				ds = append(ds, strings.TrimPrefix(d, top+"/"))
			}
		}
		dirs = ds
	}
	if _, ok := in[store.DeckPath]; !ok {
		var md []string
		for k := range in {
			if !strings.Contains(k, "/") && strings.EqualFold(path.Ext(k), ".md") {
				md = append(md, k)
			}
		}
		if len(md) != 1 {
			return ProjectFiles{}, ErrNoDeck
		}
		in[store.DeckPath] = in[md[0]]
		delete(in, md[0])
	}
	// a folder that holds a file is not an empty folder
	var empty []string
	for _, d := range dirs {
		held := false
		for k := range in {
			if strings.HasPrefix(k, d+"/") {
				held = true
				break
			}
		}
		if !held && !slices.Contains(empty, d) {
			empty = append(empty, d)
		}
	}
	slices.Sort(empty)
	return ProjectFiles{Files: in, Dirs: empty}, nil
}

// commonTop is the one folder every file is in, "" if there is none.
func commonTop(in map[string][]byte, dirs []string) string {
	top := ""
	for k := range in {
		first, _, ok := strings.Cut(k, "/")
		if !ok || (top != "" && first != top) {
			return ""
		}
		top = first
	}
	for _, d := range dirs {
		if first, _, _ := strings.Cut(d, "/"); first != top {
			return ""
		}
	}
	return top
}

// ImportFolder reads a project from files picked as a folder (the File System
// Access API's directory, or an <input webkitdirectory>): the same rules as
// a .zip, paths relative to the folder.
func ImportFolder(files map[string][]byte) (ProjectFiles, error) {
	in := map[string][]byte{}
	var total int
	for name, b := range files {
		if skipEntry(name) {
			continue
		}
		c, err := store.CleanPath(name)
		if err != nil {
			return ProjectFiles{}, err
		}
		if len(in) >= store.MaxFiles {
			return ProjectFiles{}, fmt.Errorf("a project holds at most %d files", store.MaxFiles)
		}
		if len(b) > store.MaxFileSize {
			return ProjectFiles{}, fmt.Errorf("%s is %d KiB: a file is at most %d KiB", c, len(b)>>10, store.MaxFileSize>>10)
		}
		if total += len(b); total > store.MaxProjectSize {
			return ProjectFiles{}, fmt.Errorf("the project would be over %d MiB", store.MaxProjectSize>>20)
		}
		in[c] = b
	}
	return projectFrom(in, nil)
}
