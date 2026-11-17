package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestNoIconSetInTheModules: neither the page's module nor the export engine (nor
// the playground package under them) depends on go-icons/iconoir or on
// go-widgets/toolkit/iconoir. That package embeds the whole set (about 1500 SVGs
// in one embed.FS), which a linker cannot cut down: 0.25 MiB of playground.wasm
// gzipped, for about fifty icons drawn. The playground's icons are the subset in
// package icons (cmd/iconsgen); go-widgets/toolkit v0.379.0 no longer imports the set.
func TestNoIconSetInTheModules(t *testing.T) {
	forbidden := []string{"github.com/go-icons/iconoir", "github.com/go-widgets/toolkit/iconoir"}
	for _, c := range []struct {
		pkg  string
		tags string
	}{
		{"./cmd/playground-wasm", ModuleTags},
		{"./cmd/playground-export", ExportTags},
		{".", ""},
	} {
		args := []string{"list", "-deps"}
		if c.tags != "" {
			args = append(args, "-tags", c.tags)
		}
		cmd := exec.Command("go", append(args, c.pkg)...)
		cmd.Dir = "../.."
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
		var out, errb bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &errb
		if err := cmd.Run(); err != nil {
			t.Fatalf("go list -deps %s: %v\n%s", c.pkg, err, errb.String())
		}
		deps := strings.Fields(out.String())
		// the planted positives: the list is a real one
		var sawToolkit, sawIcons bool
		for _, d := range deps {
			sawToolkit = sawToolkit || d == "github.com/go-widgets/toolkit"
			sawIcons = sawIcons || d == "github.com/go-slides/go-slides.github.io/playground/icons"
			for _, f := range forbidden {
				if d == f || strings.HasPrefix(d, f+"/") {
					t.Errorf("%s depends on %s", c.pkg, d)
				}
			}
		}
		// (the export engine draws no interface: it links neither)
		if c.pkg != "./cmd/playground-export" {
			if !sawToolkit {
				t.Errorf("go list -deps %s names no go-widgets/toolkit: the check sees nothing", c.pkg)
			}
			if !sawIcons {
				t.Errorf("%s does not use the icons subset", c.pkg)
			}
		}
	}
}
