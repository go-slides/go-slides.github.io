// Package fontsubset cuts a TrueType face down to the glyphs of a set of runes,
// keeping a character map so that the face can be looked up by rune. The
// playground uses it at build time to ship, for Ukrainian, Russian, Chinese,
// Japanese and Korean, a face of some hundreds of kilobytes instead of the
// 2 to 17 MB of the whole Noto Sans: opentype's SubsetTrueType renumbers the
// glyphs and leaves a character map that is only enough to parse, which is what
// a PDF needs and a canvas does not.
package fontsubset

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"

	"github.com/go-opentype/opentype"
)

// Subset returns ttf reduced to the glyphs of runes (those the face has), with
// a format 12 character map from each of them to its glyph.
func Subset(ttf []byte, runes []rune) ([]byte, error) {
	f, err := opentype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	seen := map[rune]opentype.GlyphIndex{}
	var gids []opentype.GlyphIndex
	for _, r := range runes {
		if g, ok := f.GlyphIndex(r); ok && g != 0 {
			if _, dup := seen[r]; !dup {
				seen[r] = g
				gids = append(gids, g)
			}
		}
	}
	if len(gids) == 0 {
		return nil, errors.New("fontsubset: the face has none of the runes")
	}
	data, oldToNew, err := f.SubsetTrueType(gids)
	if err != nil {
		return nil, err
	}
	m := make(map[rune]uint16, len(seen))
	for r, g := range seen {
		m[r] = uint16(oldToNew[g])
	}
	return withCmap(data, cmap12(m))
}

// cmap12 is a 'cmap' table with the one format 12 subtable, for (3, 10).
func cmap12(m map[rune]uint16) []byte {
	rs := make([]rune, 0, len(m))
	for r := range m {
		rs = append(rs, r)
	}
	sort.Slice(rs, func(i, j int) bool { return rs[i] < rs[j] })
	type group struct{ start, end, gid uint32 }
	var gs []group
	for _, r := range rs {
		g := uint32(m[r])
		if n := len(gs); n > 0 && gs[n-1].end+1 == uint32(r) && gs[n-1].gid+(gs[n-1].end-gs[n-1].start)+1 == g {
			gs[n-1].end = uint32(r)
			continue
		}
		gs = append(gs, group{uint32(r), uint32(r), g})
	}
	sub := make([]byte, 16+12*len(gs))
	binary.BigEndian.PutUint16(sub[0:], 12)
	binary.BigEndian.PutUint32(sub[4:], uint32(len(sub)))
	binary.BigEndian.PutUint32(sub[12:], uint32(len(gs)))
	for i, g := range gs {
		binary.BigEndian.PutUint32(sub[16+12*i:], g.start)
		binary.BigEndian.PutUint32(sub[20+12*i:], g.end)
		binary.BigEndian.PutUint32(sub[24+12*i:], g.gid)
	}
	t := make([]byte, 12, 12+len(sub))
	binary.BigEndian.PutUint16(t[2:], 1)  // one encoding record
	binary.BigEndian.PutUint16(t[4:], 3)  // Windows
	binary.BigEndian.PutUint16(t[6:], 10) // Unicode full repertoire
	binary.BigEndian.PutUint32(t[8:], 12) // its subtable follows the record
	return append(t, sub...)
}

// withCmap rewrites the sfnt's table directory with cmap in place of its own.
func withCmap(sfnt, cmap []byte) ([]byte, error) {
	if len(sfnt) < 12 {
		return nil, errors.New("fontsubset: short sfnt")
	}
	n := int(binary.BigEndian.Uint16(sfnt[4:]))
	if len(sfnt) < 12+16*n {
		return nil, errors.New("fontsubset: short table directory")
	}
	tables := map[string][]byte{"cmap": cmap}
	for i := 0; i < n; i++ {
		rec := sfnt[12+16*i:]
		tag := string(rec[:4])
		off, ln := int(binary.BigEndian.Uint32(rec[8:])), int(binary.BigEndian.Uint32(rec[12:]))
		if off < 0 || ln < 0 || off+ln > len(sfnt) {
			return nil, fmt.Errorf("fontsubset: table %q out of range", tag)
		}
		if tag != "cmap" {
			tables[tag] = sfnt[off : off+ln]
		}
	}
	tags := make([]string, 0, len(tables))
	for t := range tables {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	nt := len(tags)
	out := make([]byte, 12+16*nt)
	copy(out, sfnt[:4])
	binary.BigEndian.PutUint16(out[4:], uint16(nt))
	pow := 1
	log := 0
	for pow*2 <= nt {
		pow *= 2
		log++
	}
	binary.BigEndian.PutUint16(out[6:], uint16(pow*16))
	binary.BigEndian.PutUint16(out[8:], uint16(log))
	binary.BigEndian.PutUint16(out[10:], uint16(nt*16-pow*16))
	headAt := -1
	for i, tag := range tags {
		b := tables[tag]
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		rec := out[12+16*i:]
		copy(rec, tag)
		if tag == "head" && len(b) >= 12 {
			b = append([]byte(nil), b...)
			binary.BigEndian.PutUint32(b[8:], 0) // checkSumAdjustment, set below
			headAt = len(out)
		}
		binary.BigEndian.PutUint32(rec[4:], checksum(b))
		binary.BigEndian.PutUint32(rec[8:], uint32(len(out)))
		binary.BigEndian.PutUint32(rec[12:], uint32(len(b)))
		out = append(out, b...)
	}
	for len(out)%4 != 0 {
		out = append(out, 0)
	}
	if headAt >= 0 {
		binary.BigEndian.PutUint32(out[headAt+8:], 0xB1B0AFBA-checksum(out))
	}
	return out, nil
}

// checksum is the sum of the big-endian 32-bit words of b (zero-padded).
func checksum(b []byte) uint32 {
	var s uint32
	for i := 0; i < len(b); i += 4 {
		var w [4]byte
		copy(w[:], b[i:])
		s += binary.BigEndian.Uint32(w[:])
	}
	return s
}
