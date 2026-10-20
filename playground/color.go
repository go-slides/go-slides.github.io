package playground

import (
	"strconv"
	"strings"

	"github.com/go-widgets/toolkit"
)

// ParseColor reads "#rrggbb", "#rgb" or "rgb(r, g, b)" — what
// getComputedStyle returns for a custom property holding a colour.
func ParseColor(s string) (toolkit.RGBA, bool) {
	s = strings.TrimSpace(s)
	if h, ok := strings.CutPrefix(s, "#"); ok {
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		if len(h) != 6 {
			return toolkit.RGBA{}, false
		}
		n, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return toolkit.RGBA{}, false
		}
		return toolkit.RGB(uint8(n>>16), uint8(n>>8), uint8(n)), true
	}
	if i := strings.IndexByte(s, '('); i >= 0 && strings.HasSuffix(s, ")") {
		f := strings.FieldsFunc(s[i+1:len(s)-1], func(r rune) bool { return r == ',' || r == ' ' || r == '/' })
		if len(f) < 3 {
			return toolkit.RGBA{}, false
		}
		var c [3]uint8
		for k := range 3 {
			x, err := strconv.ParseFloat(f[k], 64)
			if err != nil {
				return toolkit.RGBA{}, false
			}
			c[k] = uint8(min(max(x, 0), 255))
		}
		return toolkit.RGB(c[0], c[1], c[2]), true
	}
	return toolkit.RGBA{}, false
}
