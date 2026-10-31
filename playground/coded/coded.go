// Package coded is an error that says what it is: a stable Code, the Args that
// fill its sentence, and its English Text. The packages that cannot know the
// playground's catalogue (store, mediafiles) return these; the playground
// explains the Code in the language of the page and falls back to Text.
package coded

import (
	"errors"
	"fmt"
)

// Error is an error with a code.
type Error struct {
	// Code is stable: it is the key of the message in the catalogue
	// ("store.maxfiles").
	Code string
	// Args fill the message's fmt verbs, in order.
	Args []any
	// Text is the English message, the one Error returns.
	Text string
}

func (e *Error) Error() string { return e.Text }

// New is an error of code, whose English Text is format filled with args.
func New(code, format string, args ...any) error {
	return &Error{Code: code, Args: args, Text: fmt.Sprintf(format, args...)}
}

// As finds the coded error in err's chain.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
