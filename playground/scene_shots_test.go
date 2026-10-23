package playground

import (
	"image"
	pngenc "image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestSceneScreenshots writes PNGs of the scene (the ribbon's tabs, light and
// dark, a menu open, a narrow window, collapsed) to the directory in
// PLAYGROUND_SHOT_DIR, for a person to look at. It writes nothing when the
// variable is unset, and refuses a directory inside a git work tree.
func TestSceneScreenshots(t *testing.T) {
	dir := os.Getenv("PLAYGROUND_SHOT_DIR")
	if dir == "" {
		t.Skip("PLAYGROUND_SHOT_DIR not set")
	}
	for d := dir; d != filepath.Dir(d); d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			t.Fatalf("%s is inside the git work tree %s", dir, d)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	shot := func(name string, s *Scene) {
		w, h := s.Size()
		buf := make([]byte, 4*w*h)
		s.Draw(buf)
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		copy(img.Pix, buf)
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := pngenc.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	}
	for scheme, pal := range map[string]Palette{"light": LightPalette, "dark": DarkPalette} {
		for _, tab := range ribbonTabs {
			s, _, _, _ := newScene(t)
			s.SetPalette(pal)
			s.Resize(1280, 640)
			gotoTab(t, s, tab)
			shot(scheme+"-"+tab, s)
		}
		s, _, _, _ := newScene(t)
		s.SetPalette(pal)
		s.Resize(1280, 640)
		gotoTab(t, s, "Insert")
		clickRect(s, "background")
		r := s.Rects()["background"]
		s.Click(r.X+r.W/2, r.Y+r.H-4) // the arrow half: the split menu
		shot(scheme+"-menu", s)
		s, _, _, _ = newScene(t)
		s.SetPalette(pal)
		s.Resize(640, 640)
		shot(scheme+"-narrow", s)
		s.ribbon.Collapsed().Set(true)
		shot(scheme+"-collapsed", s)
	}
}
