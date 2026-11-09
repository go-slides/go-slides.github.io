package playground

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
	"github.com/go-slides/go-slides.github.io/playground/mediafiles"
	"github.com/go-slides/go-slides.github.io/playground/store"
)

const videoDeck = "# Demo\n\n![video loop muted](media/clip.mp4#t=0.1)\n"

// clip is a small H.264 file (go-slides' testdata/video/bframes.mp4).
func clip(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/clip.mp4")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// A video added to the project goes in media/ as bytes (not text), is named in
// the deck as an image, and may be larger than other files, up to go-slides'
// own bound; past it, or in a file that is not a video, it is refused with a
// message.
func TestVideosAreProjectMediaWithTheirOwnBound(t *testing.T) {
	vm, _, _ := openFiles(t, store.NewMemory())
	for _, name := range []string{"My Clip.mp4", "a.M4V", "b.webm"} {
		p, err := vm.AddFile("", name, clip(t))
		if err != nil || p != "media/"+name {
			t.Fatalf("%s: %q, %v", name, p, err)
		}
		if !IsMedia(p) || !IsVideo(p) || IsRecording(p) {
			t.Errorf("%s is a video medium", p)
		}
	}
	if got := MediaMarkdown("media/My Clip.mp4"); got != "![My Clip](media/My%20Clip.mp4)" {
		t.Fatalf("markdown = %q", got)
	}
	if IsVideo("x.mov") || IsVideo("x.mp4x") || IsVideo("x.png") {
		t.Fatal("only .mp4, .m4v and .webm are videos")
	}
	if files, _ := vm.textFiles(); len(files) != 0 {
		t.Fatalf("a video is not a text file of the build: %v", files)
	}

	big := make([]byte, store.MaxFileSize+1)
	if p, err := vm.AddFile("", "big.mp4", big); err != nil {
		t.Fatalf("a video over the 8 MiB of other files: %q, %v", p, err)
	}
	if _, err := vm.AddFile("", "big.png", big); err == nil {
		t.Fatal("an image over 8 MiB was kept")
	}
	if got := store.MaxFileSizeOf("media/x.mp4"); got != 64<<20 {
		t.Fatalf("a video is at most %d", got)
	}
	if got := store.MaxFileSizeOf("media/x.png"); got != store.MaxFileSize {
		t.Fatalf("an image is at most %d", got)
	}
	p := store.NewProject("x")
	if err := p.Write("media/huge.mp4", make([]byte, mediafiles.MaxVideo+1), time.Now()); err == nil || !strings.Contains(err.Error(), "at most 65536 KiB") {
		t.Fatalf("a video past 64 MiB: %v", err)
	}
}

// The renderer is sent the videos with the project: the poster is decoded from
// them, so the preview has the poster and no video.path warning; without them
// the build says it could not find the file.
func TestServeGivesTheBuildTheVideos(t *testing.T) {
	data := clip(t)
	videoKept.Lock()
	videoKept.m = map[string]Video{}
	videoKept.Unlock()

	with := Serve(Request{ID: 1, Source: videoDeck, Videos: map[string]Video{"media/clip.mp4": {Stamp: 5, Size: len(data), Data: data}}})
	page := with.Page()
	if with.Error != "" || len(with.Diagnostics) != 0 {
		t.Fatalf("%q %+v", with.Error, with.Diagnostics)
	}
	for _, want := range []string{`class="slides-video"`, `data-slides-video="media/clip.mp4"`, `poster="data:image/jpeg;base64,`} {
		if !strings.Contains(page, want) {
			t.Errorf("the preview lacks %s", want)
		}
	}

	// the next keystroke leaves the bytes out: the worker holds that version
	held := Serve(Request{ID: 2, Source: videoDeck + "\nmore\n", Videos: map[string]Video{"media/clip.mp4": {Stamp: 5, Size: len(data)}}})
	if held.Error != "" || !strings.Contains(held.Page(), `poster="data:image/jpeg;base64,`) {
		t.Fatalf("a held video: %q", held.Error)
	}
	// a version the worker was never sent is refused, not guessed
	if r := Serve(Request{ID: 3, Source: videoDeck, Videos: map[string]Video{"media/clip.mp4": {Stamp: 6, Size: len(data)}}}); r.Error != ErrVideoNotHeld.Error() {
		t.Fatalf("an unknown version: %q", r.Error)
	}
	// and a path that is not a video is refused
	if _, err := resolveVideos(map[string]Video{"media/a.png": {Stamp: 1, Size: 1, Data: []byte("x")}}); err == nil {
		t.Fatal("a non-video among the videos")
	}
	if _, err := resolveVideos(map[string]Video{"../clip.mp4": {Stamp: 1, Size: 1, Data: []byte("x")}}); err == nil {
		t.Fatal("a path out of the project")
	}
	if _, err := resolveVideos(map[string]Video{"media/clip.mp4": {Stamp: 1, Size: 1}}); !errors.Is(err, ErrVideoNotHeld) {
		t.Fatalf("a request that names a video and has no bytes: %v", err)
	}

	none := Serve(Request{ID: 4, Source: videoDeck})
	if len(none.Diagnostics) == 0 || strings.Contains(none.Page(), "poster=") {
		t.Fatalf("without the video the build says nothing: %+v", none.Diagnostics)
	}
	resolveVideos(nil)
}

// A request carries a video's bytes once per worker and version: the page
// leaves them out afterwards, and sends them again to a new worker or for a
// version that was written again or is only previewed.
func TestRequestsCarryAVideoOncePerWorker(t *testing.T) {
	data := []byte("bytes")
	r := Request{Videos: map[string]Video{"media/a.mp4": {Stamp: 7, Size: 5, Data: data}, "media/b.webm": {Stamp: -1, Size: 5, Data: data}}}
	first, held := WithoutHeldVideos(r, nil)
	if first.Videos["media/a.mp4"].Data == nil || first.Videos["media/b.webm"].Data == nil {
		t.Fatal("the first request carries the bytes")
	}
	if len(held) != 1 || held["media/a.mp4"] != (VideoVersion{7, 5}) {
		t.Fatalf("a snapshot is never held: %v", held)
	}
	second, held := WithoutHeldVideos(r, held)
	if second.Videos["media/a.mp4"].Data != nil || second.Videos["media/b.webm"].Data == nil {
		t.Fatal("the held version goes without its bytes, the snapshot with")
	}
	if r.Videos["media/a.mp4"].Data == nil {
		t.Fatal("the request the view-model keeps was changed")
	}
	r.Videos["media/a.mp4"] = Video{Stamp: 8, Size: 5, Data: data}
	third, held := WithoutHeldVideos(r, held)
	if third.Videos["media/a.mp4"].Data == nil {
		t.Fatal("a version written again is sent again")
	}
	if none, held := WithoutHeldVideos(Request{}, held); none.Videos != nil || held != nil {
		t.Fatal("a request with no video makes the worker forget them")
	}
	b, _ := json.Marshal(second)
	if strings.Contains(string(b), `"media/a.mp4":{"stamp":7,"size":5,"data"`) {
		t.Fatal("the bytes are in the JSON")
	}
}

// The view-model puts the project's videos in the request it sends, and a
// version in the preview's, never kept.
func TestViewModelSendsTheVideos(t *testing.T) {
	vm, _, tr := openFiles(t, store.NewMemory())
	data := clip(t)
	if _, err := vm.AddFile("", "clip.mp4", data); err != nil {
		t.Fatal(err)
	}
	vm.Source.Set(videoDeck)
	got := tr.sent[len(tr.sent)-1].Videos["media/clip.mp4"]
	if got.Size != len(data) || string(got.Data) != string(data) || got.Stamp < 0 {
		t.Fatalf("request video: stamp %d size %d", got.Stamp, got.Size)
	}
	if v := vm.VideoFiles(); len(v) != 1 || string(v["media/clip.mp4"]) != string(data) {
		t.Fatal("an export is sent the videos")
	}
}

// An export's deck is built with the videos, and the Doc it makes carries
// them for the PPTX and ODP: (doc.Doc.Media, JSON "media") base64.
func TestExportDocCarriesTheVideos(t *testing.T) {
	data := clip(t)
	d, err := exportdoc.Build(exportdoc.Project{DeckPath: DeckPath, Source: videoDeck, Videos: map[string][]byte{"media/clip.mp4": data}})
	if err != nil {
		t.Fatal(err)
	}
	m, ok := d.Media["media/clip.mp4"]
	if !ok || m.MIME != "video/mp4" || string(m.Data) != string(data) {
		t.Fatalf("media = %v", d.Media)
	}
	b, _ := json.Marshal(d)
	if !regexp.MustCompile(`"media":\{"media/clip.mp4":\{"mime":"video/mp4","data":"[A-Za-z0-9+/=]+"\}\}`).Match(b) {
		t.Fatalf("the Doc's JSON has no media")
	}
	if _, err := exportdoc.Build(exportdoc.Project{DeckPath: DeckPath, Source: videoDeck, Videos: map[string][]byte{"media/a.png": data}}); err == nil {
		t.Fatal("a file that is not a video among the videos")
	}
}

// The frame's script puts a blob: URL of the project's file in a video's src,
// keeps the start time, revokes the URL it replaces and gives back the path
// of a file that left; and nothing but the page's own media is let in (see
// browsertest/video.cjs for the script run in Chrome).
func TestMediaScriptMakesBlobURLsOfVideos(t *testing.T) {
	for _, want := range []string{
		`data-slides-video`, `URL.createObjectURL`, `URL.revokeObjectURL`,
		`tag==="VIDEO"`, `cur.slice(h)`, `/^(image|video|audio|font)\/`,
	} {
		if !strings.Contains(mediaScript, want) {
			t.Errorf("mediaScript lacks %q", want)
		}
	}
}

// media-src carries blob: (and only what the media of the page need), in the
// preview frame's policy and in the playground page's, as the Content Security
// Policy test of the page pins it.
func TestPoliciesAllowBlobMedia(t *testing.T) {
	if p := csp("'nonce-x'"); !strings.Contains(p, "media-src https: data: blob:;") || !strings.Contains(p, "default-src 'none'") {
		t.Fatalf("the frame's policy: %s", p)
	}
	b, err := os.ReadFile("../layouts/_default/playground.html")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`<meta http-equiv="Content-Security-Policy" content="([^"]+)">`).FindSubmatch(b)
	if m == nil {
		t.Fatal("no policy in the page")
	}
	if !strings.Contains(string(m[1]), "media-src 'self' https: data: blob:;") || !strings.Contains(string(m[1]), "default-src 'self'") {
		t.Fatalf("the page's policy: %s", m[1])
	}
}

// TestWriteVideoPreviewForBrowser writes the preview frame's document, for a
// deck with a video, and the video, into the folder VIDEO_PREVIEW_OUT names,
// for browsertest/video.cjs. It writes nothing, and passes, without it.
func TestWriteVideoPreviewForBrowser(t *testing.T) {
	out := os.Getenv("VIDEO_PREVIEW_OUT")
	if out == "" {
		t.Skip("VIDEO_PREVIEW_OUT is not set")
	}
	data := clip(t)
	r := Serve(Request{ID: 1, Source: videoDeck, Videos: map[string]Video{"media/clip.mp4": {Stamp: 1, Size: len(data), Data: data}}})
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if err := os.WriteFile(out+"/frame.html", []byte(r.Page()), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out+"/clip.mp4", data, 0o600); err != nil {
		t.Fatal(err)
	}
}
