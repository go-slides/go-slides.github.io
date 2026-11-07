package playground

import (
	"encoding/json"
	"strings"
)

// How "Remember on this device" writes a token: as a record that names the
// forge and the API origin it is for, so that it is never read back for another.
// The host (cmd/playground-wasm) keeps the string under KeptTokenKey; the codec is
// here so that it is tested natively.

// KeptTokenKey is the localStorage key of the record.
const KeptTokenKey = "goSlidesGitToken"

// LegacyTokenKey is where an earlier page kept a bare token, with no record of
// the forge it was for (a GitLab token could sit there as well as a GitHub one):
// it is never used, only removed.
const LegacyTokenKey = "goSlidesGitHubToken"

type keptRecord struct {
	V      int    `json:"v"`
	Forge  string `json:"forge"`
	Origin string `json:"origin"`
	Token  string `json:"token"`
}

// EncodeKept is the record that holds token for forge and origin.
func EncodeKept(forge, origin, token string) string {
	b, _ := json.Marshal(keptRecord{V: 1, Forge: forge, Origin: origin, Token: token})
	return string(b)
}

// DecodeKept reads a record; ok is false for anything that is not one with all
// three fields (a bare token, damaged text, another version).
func DecodeKept(s string) (forge, origin, token string, ok bool) {
	if !strings.HasPrefix(s, "{") {
		return "", "", "", false
	}
	var r keptRecord
	if json.Unmarshal([]byte(s), &r) != nil || r.V != 1 || r.Forge == "" || r.Origin == "" || r.Token == "" {
		return "", "", "", false
	}
	return r.Forge, r.Origin, r.Token, true
}
