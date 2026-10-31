package playground

import (
	"regexp"
	"strings"
)

// The Format tab's table commands. A table is the GFM kind the Marp preview renders
// (goldmark's table extension): a header row, a delimiter row of `---` cells whose
// colons set the column's alignment (`:---` left, `:---:` centre, `---:` right),
// then the body rows. Every command reads the table around the caret, changes it and
// writes it back pretty-aligned, so that the Markdown stays readable after each edit.

// Column alignments of a table.
const (
	AlignNone   byte = 0
	AlignLeft   byte = 'l'
	AlignCenter byte = 'c'
	AlignRight  byte = 'r'
)

// tableBlock is a table read from the deck: its lines, the cells of the header and
// the body (row 0 is the header), and the alignment of each column.
type tableBlock struct {
	first, last int
	rows        [][]string
	aligns      []byte
}

var delimCell = regexp.MustCompile(`^:?-+:?$`)

// splitCells is the cells of a table line, without the outer pipes; `\|` is a pipe
// in a cell.
func splitCells(line string) []string {
	s := strings.TrimSpace(line)
	r := []rune(s)
	if len(r) > 0 && r[0] == '|' {
		r = r[1:]
	}
	if len(r) > 0 && r[len(r)-1] == '|' && (len(r) < 2 || r[len(r)-2] != '\\') {
		r = r[:len(r)-1]
	}
	var cells []string
	start := 0
	for i := 0; i <= len(r); i++ {
		if i == len(r) || r[i] == '|' && (i == 0 || r[i-1] != '\\') {
			cells = append(cells, strings.TrimSpace(string(r[start:i])))
			start = i + 1
		}
	}
	return cells
}

// isDelimiterRow reports whether a line is the `---|---` row under a header.
func isDelimiterRow(line string) bool {
	if !strings.Contains(line, "|") && !delimCell.MatchString(strings.TrimSpace(line)) {
		return false
	}
	cells := splitCells(line)
	for _, c := range cells {
		if !delimCell.MatchString(c) {
			return false
		}
	}
	return len(cells) > 0
}

func alignOf(cell string) byte {
	l, r := strings.HasPrefix(cell, ":"), strings.HasSuffix(cell, ":")
	switch {
	case l && r:
		return AlignCenter
	case l:
		return AlignLeft
	case r:
		return AlignRight
	}
	return AlignNone
}

// findTable reads the table the line is in.
func findTable(lines []string, line int) (tableBlock, bool) {
	if line < 0 || line >= len(lines) || !plainLines(lines)[line] || !strings.Contains(lines[line], "|") {
		return tableBlock{}, false
	}
	top, bottom := line, line
	for top > 0 && strings.Contains(lines[top-1], "|") && !isBlank(lines[top-1]) {
		top--
	}
	for bottom+1 < len(lines) && strings.Contains(lines[bottom+1], "|") && !isBlank(lines[bottom+1]) {
		bottom++
	}
	for d := top + 1; d <= bottom; d++ {
		if !isDelimiterRow(lines[d]) {
			continue
		}
		header := splitCells(lines[d-1])
		delim := splitCells(lines[d])
		if len(header) != len(delim) || line < d-1 {
			continue
		}
		tb := tableBlock{first: d - 1, last: bottom, rows: [][]string{header}}
		for _, c := range delim {
			tb.aligns = append(tb.aligns, alignOf(c))
		}
		for b := d + 1; b <= bottom; b++ {
			cells := splitCells(lines[b])
			for len(cells) < len(header) {
				cells = append(cells, "")
			}
			tb.rows = append(tb.rows, cells[:len(header)])
		}
		return tb, true
	}
	return tableBlock{}, false
}

// render writes the table as aligned lines, and says where each cell's text starts
// (rune columns) in the line of its row (0 the header, k the k-th body row).
func (tb tableBlock) render() (lines []string, starts [][]int) {
	n := len(tb.aligns)
	width := make([]int, n)
	for c := 0; c < n; c++ {
		width[c] = 3
		for _, row := range tb.rows {
			width[c] = max(width[c], len([]rune(row[c])))
		}
	}
	pad := func(text string, c int) string {
		gap := width[c] - len([]rune(text))
		switch tb.aligns[c] {
		case AlignRight:
			return strings.Repeat(" ", gap) + text
		case AlignCenter:
			return strings.Repeat(" ", gap/2) + text + strings.Repeat(" ", gap-gap/2)
		}
		return text + strings.Repeat(" ", gap)
	}
	delim := func(c int) string {
		w := width[c]
		switch tb.aligns[c] {
		case AlignLeft:
			return ":" + strings.Repeat("-", w-1)
		case AlignRight:
			return strings.Repeat("-", w-1) + ":"
		case AlignCenter:
			return ":" + strings.Repeat("-", w-2) + ":"
		}
		return strings.Repeat("-", w)
	}
	row := func(cell func(c int) string) string {
		parts := make([]string, n)
		for c := range parts {
			parts[c] = cell(c)
		}
		return "| " + strings.Join(parts, " | ") + " |"
	}
	cellStarts := func(row []string) []int {
		out := make([]int, n)
		x := 2
		for c := 0; c < n; c++ {
			gap := width[c] - len([]rune(row[c]))
			lead := 0
			switch tb.aligns[c] {
			case AlignRight:
				lead = gap
			case AlignCenter:
				lead = gap / 2
			}
			out[c] = x + lead
			x += width[c] + 3
		}
		return out
	}
	for i, r := range tb.rows {
		r := r
		lines = append(lines, row(func(c int) string { return pad(r[c], c) }))
		starts = append(starts, cellStarts(r))
		if i == 0 {
			lines = append(lines, row(delim))
		}
	}
	return lines, starts
}

// where is the table cell a caret is in: its row (0 the header, k the k-th body
// row, -1 the delimiter row) and its column.
func (tb tableBlock) where(lines []string, caret Pos) (row, col int) {
	switch {
	case caret.Line == tb.first:
		row = 0
	case caret.Line == tb.first+1:
		row = -1
	default:
		row = caret.Line - tb.first - 1
	}
	r := []rune(lines[caret.Line])
	pipes, lead := 0, strings.HasPrefix(strings.TrimSpace(lines[caret.Line]), "|")
	for i := 0; i < caret.Col && i < len(r); i++ {
		if r[i] == '|' && (i == 0 || r[i-1] != '\\') {
			pipes++
		}
	}
	if lead {
		pipes--
	}
	return row, min(max(pipes, 0), len(tb.aligns)-1)
}

// commit writes a table back over the lines it came from, the caret in cell
// (row, col): row 0 the header, k the k-th body row.
func (tb tableBlock) commit(lines []string, row, col int) Change {
	out, starts := tb.render()
	line := tb.first + row
	if row > 0 {
		line++ // the delimiter row sits between
	}
	end := len([]rune(lines[tb.last]))
	c := Change{From: Pos{tb.first, 0}, To: Pos{tb.last, end}, Text: strings.Join(out, "\n")}
	cell := tb.rows[row][col]
	c.Caret = Pos{line, starts[row][col] + len([]rune(cell))}
	return c
}

func emptyRow(n int) []string { return make([]string, n) }

// TableState says whether the caret is in a table, and the alignment of its column.
func TableState(src string, caret Pos) (in bool, align byte) {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	tb, ok := findTable(lines, caret.Line)
	if !ok {
		return false, AlignNone
	}
	_, col := tb.where(lines, caret)
	return true, tb.aligns[col]
}

// tableEdit reads the table at the caret, lets f change it (f returns the cell the
// caret goes to), and writes it back aligned.
func tableEdit(src string, caret Pos, f func(tb *tableBlock, row, col int) (int, int, bool)) (Change, bool) {
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	tb, ok := findTable(lines, caret.Line)
	if !ok {
		return Change{From: caret, To: caret, Caret: caret}, false
	}
	row, col := tb.where(lines, caret)
	r, c, ok := f(&tb, row, col)
	if !ok {
		return Change{From: caret, To: caret, Caret: caret}, false
	}
	return tb.commit(lines, r, c), true
}

// TableAddRow adds an empty body row under the caret's row (above it when below is
// false; nothing goes above the header).
func TableAddRow(src string, caret Pos, below bool) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) {
		at := row + 1
		if !below {
			at = row
		}
		at = max(at, 1)
		if row < 0 {
			at = 1
		}
		tb.rows = append(tb.rows[:at], append([][]string{emptyRow(len(tb.aligns))}, tb.rows[at:]...)...)
		return at, max(col, 0), true
	})
}

// TableDeleteRow removes the caret's body row.
func TableDeleteRow(src string, caret Pos) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) {
		if row < 1 {
			return 0, 0, false // the header and the delimiter stay
		}
		tb.rows = append(tb.rows[:row], tb.rows[row+1:]...)
		return min(row, len(tb.rows)-1), col, true
	})
}

// TableAddColumn adds an empty column right of the caret's column (left of it when
// right is false).
func TableAddColumn(src string, caret Pos, right bool) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) {
		at := col
		if right {
			at = col + 1
		}
		for i, r := range tb.rows {
			cell := ""
			if i == 0 {
				cell = T(KTableColumn, at+1)
			}
			tb.rows[i] = append(r[:at], append([]string{cell}, r[at:]...)...)
		}
		tb.aligns = append(tb.aligns[:at], append([]byte{AlignNone}, tb.aligns[at:]...)...)
		return max(row, 0), at, true
	})
}

// TableDeleteColumn removes the caret's column; a table keeps at least one.
func TableDeleteColumn(src string, caret Pos) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) {
		if len(tb.aligns) < 2 {
			return 0, 0, false
		}
		for i, r := range tb.rows {
			tb.rows[i] = append(r[:col], r[col+1:]...)
		}
		tb.aligns = append(tb.aligns[:col], tb.aligns[col+1:]...)
		return max(row, 0), min(col, len(tb.aligns)-1), true
	})
}

// TableAlign sets the alignment of the caret's column, or clears it when the column
// has that alignment already.
func TableAlign(src string, caret Pos, a byte) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) {
		if tb.aligns[col] == a {
			a = AlignNone
		}
		tb.aligns[col] = a
		return max(row, 0), col, true
	})
}

// TableFormat writes the table around the caret aligned, as it is.
func TableFormat(src string, caret Pos) (Change, bool) {
	return tableEdit(src, caret, func(tb *tableBlock, row, col int) (int, int, bool) { return max(row, 0), col, true })
}

// Table size limits of InsertTable.
const (
	MaxTableCols = 12
	MaxTableRows = 30
)

// InsertTable writes a table of cols columns and rows rows (the header counts) on
// lines of its own after the caret's line (in place, on an empty line; after the
// table, inside one), the header's first cell selected for typing over.
func InsertTable(src string, caret Pos, sel Range, cols, rows int) Change {
	cols, rows = min(max(cols, 1), MaxTableCols), min(max(rows, 2), MaxTableRows)
	lines := splitLines(src)
	caret = clampPos(lines, caret)
	if s := normalize(lines, sel); !s.Empty() {
		caret = s.End
	}
	tb := tableBlock{aligns: make([]byte, cols)}
	header := make([]string, cols)
	for c := range header {
		header[c] = T(KTableColumn, c+1)
	}
	tb.rows = append(tb.rows, header)
	for r := 1; r < rows; r++ {
		tb.rows = append(tb.rows, emptyRow(cols))
	}
	text, starts := tb.render()
	table := strings.Join(text, "\n")
	line := caret.Line
	if t, ok := findTable(lines, line); ok {
		line = t.last
	}
	var c Change
	first := line
	if isBlank(lines[line]) {
		c = Change{From: Pos{line, 0}, To: Pos{line, 0}, Text: table + "\n"}
	} else {
		after := ""
		if line+1 < len(lines) && !isBlank(lines[line+1]) {
			after = "\n"
		}
		end := Pos{line, len([]rune(lines[line]))}
		c = Change{From: end, To: end, Text: "\n\n" + table + after}
		first = line + 2
	}
	name := len([]rune(header[0]))
	c.Select = &Range{Pos{first, starts[0][0]}, Pos{first, starts[0][0] + name}}
	c.Caret = c.Select.End
	return c
}
