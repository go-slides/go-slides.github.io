package fontsubset

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// The characters a person writes a deck in, per language, are the repertoires
// of the national standards: GB 2312 (6,763 hanzi and the punctuation) for
// Simplified Chinese, JIS X 0208 (kana, 6,355 kanji) for Japanese, KS X 1001
// (2,350 syllables) for Korean. They are read from golang.org/x/text's decoders
// rather than listed here, by decoding every two-byte code of the standard.

// GB2312 is the repertoire of GB 2312.
func GB2312() []rune { return decodePairs(simplifiedchinese.GBK, 0xA1, 0xF7, 0xA1, 0xFE) }

// JISX0208 is the repertoire of JIS X 0208, through Shift_JIS.
func JISX0208() []rune {
	r := decodePairs(japanese.ShiftJIS, 0x81, 0xEA, 0x40, 0xFC)
	// the half-width katakana are single bytes
	for b := 0xA1; b <= 0xDF; b++ {
		if s, err := japanese.ShiftJIS.NewDecoder().Bytes([]byte{byte(b)}); err == nil {
			r = append(r, []rune(string(s))...)
		}
	}
	return r
}

// KSX1001 is the repertoire of KS X 1001 through EUC-KR, without its 4,888
// hanja: the syllables and the symbols a Korean deck is written in. (A hanja
// is rare enough, and a face without it would cost four times as much.)
func KSX1001() []rune {
	var out []rune
	for _, r := range decodePairs(korean.EUCKR, 0xA1, 0xFD, 0xA1, 0xFE) {
		if r < 0x4E00 || r > 0x9FFF {
			out = append(out, r)
		}
	}
	return out
}

// decodePairs decodes each two-byte code with lead in [l0,l1] and trail in
// [t0,t1] and keeps those that are a character.
func decodePairs(e encoding.Encoding, l0, l1, t0, t1 int) []rune {
	dec := e.NewDecoder()
	var out []rune
	for l := l0; l <= l1; l++ {
		for t := t0; t <= t1; t++ {
			s, err := dec.Bytes([]byte{byte(l), byte(t)})
			if err != nil {
				dec.Reset()
				continue
			}
			rs := []rune(string(s))
			if len(rs) == 1 && rs[0] != 0xFFFD && rs[0] > 0x7F {
				out = append(out, rs[0])
			}
		}
	}
	return out
}

// Range is the runes from lo to hi.
func Range(lo, hi rune) []rune {
	out := make([]rune, 0, hi-lo+1)
	for r := lo; r <= hi; r++ {
		out = append(out, r)
	}
	return out
}

// Latin is Basic Latin, Latin-1, Latin Extended-A and the General Punctuation
// block: what every language's text uses besides its own script.
func Latin() []rune {
	return append(append(append(Range(0x20, 0x7E), Range(0xA0, 0x17F)...), Range(0x2000, 0x206F)...), Range(0x20A0, 0x20BF)...)
}

// Cyrillic is the Cyrillic block, with the supplement Ukrainian and Russian use.
func Cyrillic() []rune { return append(Range(0x400, 0x4FF), Range(0x500, 0x52F)...) }
