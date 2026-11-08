package playground

import (
	"os"
	"testing"

	"github.com/go-slides/slides/editor"
)

// TestWritePreviewPageForBrowser writes the preview frame's document, for a
// deck with a recording, to the file CAST_PREVIEW_OUT names, for
// browsertest/cast.cjs to load in a sandboxed iframe under the playground
// page's policy. It writes nothing, and passes, without the variable.
func TestWritePreviewPageForBrowser(t *testing.T) {
	out := os.Getenv("CAST_PREVIEW_OUT")
	if out == "" {
		t.Skip("CAST_PREVIEW_OUT is not set")
	}
	cast := "{\"version\":2,\"width\":40,\"height\":6}\n[0.2,\"o\",\"\\u001b[32mhello\\u001b[0m\\r\\n\"]\n[1.2,\"o\",\"\\u001b[1;31mworld\\u001b[0m\"]\n[2.0,\"o\",\"\\r\\nthe end\"]\n"
	res := editor.Render(editor.Request{
		Path: "deck.md", Source: "# Recording\n\n![A demo](media/demo.cast)\n",
		Files: map[string]string{"media/demo.cast": cast},
	})
	if err := os.WriteFile(out, []byte(overviewPage(res.HTML, SlideTop{I: -1})), 0o600); err != nil {
		t.Fatal(err)
	}
}
