// Package icons serves the Iconoir glyphs the playground draws, and only those.
//
// go-icons/iconoir embeds its whole set (about 1500 SVGs) in one embed.FS, and
// a linker cannot drop part of an embedded file system: importing it puts every
// icon in playground.wasm, 0.25 to 0.3 MiB gzipped, for about fifty drawn. The
// glyphs here are generated from the names the playground's sources use
// (cmd/iconsgen), unmodified; the Iconoir artwork is MIT, see ICONOIR-LICENSE.
package icons

//go:generate go run ../cmd/iconsgen --dir .. --out icons_gen.go

// Icon returns the SVG document of the Iconoir icon named name, or "" when the
// playground does not use it (add it to the sources and run go generate).
func Icon(name string) string { return svg[name] }

// Len is how many icons there are, for a test that guards the subset.
func Len() int { return len(svg) }
