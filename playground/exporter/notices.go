package exporter

import (
	"regexp"

	"github.com/go-slides/go-slides.github.io/playground/coded"
)

// Notices are the export's warning lines (Result.Warnings, the lines the
// slides command prints on its standard error) as coded errors, so that the
// page says them in its own language: the code is the key of the message
// ("err.export.native"), the Args fill it (all strings), and the Text is the
// line as the export wrote it, the fallback. A line this does not know is
// "export.other".
//
// The lines are those of go-slides' export package (svganim.go, svgemit.go):
// each animated SVG image is carried natively, as an animated GIF, kept still
// or, in a PDF, drawn at its initial state; a local file the options do not
// allow, or that does not exist, is named. The technical reason an image could
// not be carried natively stays as the export wrote it.
func Notices(lines []string) []*coded.Error {
	out := make([]*coded.Error, 0, len(lines))
	for _, l := range lines {
		out = append(out, notice(l))
	}
	return out
}

var (
	slidesRe  = `\(slides? ([^)]*)\)`
	nativeRe  = regexp.MustCompile(`^svg-animation: (.+?) ` + slidesRe + `: animated natively in (\w+): (\d+) elements? \(([^)]*)\)(.*)$`)
	framesRe  = regexp.MustCompile(`^svg-animation: (.+?) ` + slidesRe + `: animated as frames in (\w+), an animated GIF: (.*)$`)
	initialRe = regexp.MustCompile(`^svg-animation: (.+?) ` + slidesRe + ` is animated, which`)
	gifRe     = regexp.MustCompile(`^svg-animation: (.+?) is an animated GIF of (\d+) frames at (\d+) fps, (\d+x\d+) pixels, (\d+) KiB: (.*)$`)
	stillRe   = regexp.MustCompile(`^svg-animation: (.+?) is kept as a still picture[^:]*: (.*)$`)
	blockedRe = regexp.MustCompile(`^local-file-blocked: "(.*?)" ` + slidesRe + ` is a local file`)
	missingRe = regexp.MustCompile(`^local-file-missing: "(.*?)" ` + slidesRe + ` does not exist`)
)

func notice(l string) *coded.Error {
	mk := func(code string, args ...string) *coded.Error {
		a := make([]any, len(args))
		for i, s := range args {
			a[i] = s
		}
		return &coded.Error{Code: "export." + code, Args: a, Text: l}
	}
	if m := nativeRe.FindStringSubmatch(l); m != nil {
		return mk("native", m[1], m[2], m[3], m[4], m[5], m[6])
	}
	if m := framesRe.FindStringSubmatch(l); m != nil {
		return mk("frames", m[1], m[2], m[3], m[4])
	}
	if m := initialRe.FindStringSubmatch(l); m != nil {
		return mk("initial", m[1], m[2])
	}
	if m := gifRe.FindStringSubmatch(l); m != nil {
		return mk("gif", m[1], m[2], m[3], m[4], m[5], m[6])
	}
	if m := stillRe.FindStringSubmatch(l); m != nil {
		return mk("still", m[1], m[2])
	}
	if m := blockedRe.FindStringSubmatch(l); m != nil {
		return mk("blocked", m[1], m[2])
	}
	if m := missingRe.FindStringSubmatch(l); m != nil {
		return mk("missing", m[1], m[2])
	}
	return mk("other", l)
}
