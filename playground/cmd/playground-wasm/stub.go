// Copyright (c) the go-slides authors.
// SPDX-License-Identifier: BSD-3-Clause

//go:build !(js && wasm)

// The playground runs in a browser only; natively this command is empty, so
// that `go vet ./...` and `go test ./...` cover the module on any machine.
package main

func main() {}
