package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-tex/math"
)

// MathFontFile and MathFontLicense are where writeMathFont puts the font the
// formulas are typeset with, and its licence, under out/fonts/. The module is
// built with go-slides' tag slides_nomathfont (ModuleTags), so it links no font
// data (651 KB gzipped less); the render worker fetches this file the first
// time a render has a formula (cmd/playground-wasm/lazyfetch.go, which asks for
// "fonts/" + MathFontFile).
const (
	MathFontFile    = "stix-math.otf"
	MathFontLicense = "stix-math.OFL.txt"
)

// mathFontOFL is STIX Two Math's licence, the SIL Open Font License 1.1, verbatim
// from go-tex/math's STIXTwoMath-OFL.txt (the module the bytes come from). A
// copy of a font is a redistribution: it carries its licence.
//
//go:embed STIXTwoMath-OFL.txt
var mathFontOFL []byte

// writeMathFont writes STIX Two Math under dir: the very bytes go-slides'
// embedded provider (mathfont/embedded) serves, which are go-tex/math's
// DefaultFont, so that a formula in the browser is typeset by the font the
// `slides` command typesets with. It returns the font's size and SHA-256.
func writeMathFont(dir string) (digest, error) {
	b := math.DefaultFont()
	if len(b) == 0 {
		return digest{}, fmt.Errorf("go-tex/math has no default font in this build")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return digest{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, MathFontFile), b, 0o644); err != nil {
		return digest{}, err
	}
	if err := os.WriteFile(filepath.Join(dir, MathFontLicense), mathFontOFL, 0o644); err != nil {
		return digest{}, err
	}
	return digestOf(b), nil
}
