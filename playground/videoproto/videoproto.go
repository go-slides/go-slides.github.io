// Package videoproto is the message protocol of a video worker: a program that
// reads video files for a page that must not link a container reader or a
// decoder (the playground's WebAssembly module). It is two typed requests,
//
//	poster(file, t)  -> a JPEG, its width and height
//	probe(file)      -> duration, codec, size
//
// answered with the coded refusals of go-slides' video package (video.limit,
// video.syntax, video.codec, video.unavailable). The messages are plain data:
// the transport is the host's (a MessagePort between two Web Workers in the
// playground; a function call in the tests), and nothing here imports syscall/js,
// a decoder or go-slides beyond its Limits type, so another host can reuse it.
//
// A message is a JSON header and, beside it, a binary payload: the file in a
// request, the picture in a response. The payload is never inside the JSON (a
// file is up to 64 MiB). A worker keeps the files it was sent, by Key, so that
// the next requests for the same file (a poster at another time, a keystroke
// later) do not carry it again; a request for a file it does not hold is
// answered NotHeld, and the client sends it again with the file.
package videoproto

import (
	"hash/crc64"
	"strconv"

	"github.com/go-slides/slides/video"
)

// The operations.
const (
	OpPoster = "poster"
	OpProbe  = "probe"
)

// CodeNotHeld is the code of an answer that is not a refusal of the file: the
// worker does not hold the file the request names. It is never shown.
const CodeNotHeld = "video.notheld"

// Request asks a worker to read a file.
type Request struct {
	// ID pairs the answer with the request.
	ID int `json:"id"`
	// Op is OpPoster or OpProbe.
	Op string `json:"op"`
	// Key names the version of the file (Fingerprint).
	Key string `json:"key"`
	// HasData says the payload is the file; otherwise the worker's copy of the
	// file named by Key is read.
	HasData bool `json:"hasData,omitempty"`
	// Format is "mp4" or "webm", as go-slides sniffed it (informative: the
	// worker sniffs again).
	Format string `json:"format,omitempty"`
	// T is the poster's time, in seconds.
	T float64 `json:"t,omitempty"`
	// Limits are the bounds the host asks for; the worker never exceeds
	// video.DefaultLimits (Clamp).
	Limits video.Limits `json:"limits"`
}

// Response answers a Request. The poster's picture is the payload.
type Response struct {
	ID int `json:"id"`
	// Code and Error are a refusal: a video.* code and its message.
	Code  string `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
	// Probe's answer: the container ("mp4", "webm"), the video track, and
	// whether a poster can be decoded for it.
	Format    string  `json:"format,omitempty"`
	MIME      string  `json:"mime,omitempty"`
	Codec     string  `json:"codec,omitempty"`
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	Duration  float64 `json:"duration,omitempty"`
	Decodable bool    `json:"decodable,omitempty"`
	// Image is the type of the payload of a poster's answer ("image/jpeg"); its
	// size is Width x Height for a poster.
	Image string `json:"image,omitempty"`
}

// OK reports whether the answer is a refusal's or not.
func (r Response) OK() bool { return r.Code == "" }

// Fingerprint names a version of a file: its length and a 64-bit checksum of
// every byte (CRC-64/ECMA), which costs a few milliseconds per 10 MiB and which
// a trimmed or re-encoded file does not keep.
func Fingerprint(data []byte) string {
	return strconv.Itoa(len(data)) + "-" + strconv.FormatUint(crc64.Checksum(data, crcTable), 16)
}

var crcTable = crc64.MakeTable(crc64.ECMA)

// Clamp is the limits a worker applies for a request's: each bound of ask is
// kept when it is below the worker's own (video.DefaultLimits), and replaced
// by it otherwise or when it is not positive. A host can lower the bounds, not
// raise them.
func Clamp(ask video.Limits) video.Limits {
	d := video.DefaultLimits
	c := func(a, max int) int {
		if a <= 0 || a > max {
			return max
		}
		return a
	}
	return video.Limits{
		MaxBytes: c(ask.MaxBytes, d.MaxBytes), MaxBoxes: c(ask.MaxBoxes, d.MaxBoxes), MaxDepth: c(ask.MaxDepth, d.MaxDepth),
		MaxWidth: c(ask.MaxWidth, d.MaxWidth), MaxHeight: c(ask.MaxHeight, d.MaxHeight),
		MaxSamples: c(ask.MaxSamples, d.MaxSamples), MaxTrackSamples: c(ask.MaxTrackSamples, d.MaxTrackSamples),
		MaxSampleBytes: c(ask.MaxSampleBytes, d.MaxSampleBytes), PosterWidth: c(ask.PosterWidth, d.PosterWidth),
	}
}
