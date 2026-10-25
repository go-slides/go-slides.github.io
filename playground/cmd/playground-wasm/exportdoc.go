// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build js && wasm

package main

import (
	"encoding/json"

	"github.com/go-slides/go-slides.github.io/playground/exportdoc"
)

// docRequest asks a worker of this module to render the deck for an export
// (exportdoc.Build): the export engine's module carries no renderer.
type docRequest struct {
	ExportDoc bool `json:"exportDoc"`
	ID        int  `json:"id"`
	exportdoc.Project
}

// docResponse is the deck rendered, as go-slides' doc.Doc in JSON, or why
// not.
type docResponse struct {
	ID    int             `json:"id"`
	Doc   json.RawMessage `json:"doc,omitempty"`
	Error string          `json:"error,omitempty"`
}

// serveDoc renders the deck of r: the answer, as the JSON string the
// worker posts.
func serveDoc(r docRequest) string {
	res := docResponse{ID: r.ID}
	d, err := exportdoc.Build(r.Project)
	if err != nil {
		res.Error = err.Error()
	} else if res.Doc, err = json.Marshal(d); err != nil {
		res.Error = err.Error()
	}
	b, _ := json.Marshal(res)
	return string(b)
}
