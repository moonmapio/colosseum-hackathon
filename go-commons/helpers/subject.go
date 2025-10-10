package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

func SafeToken(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return s
}

func RandID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// just min y [a-z0-9_-]
var reBad = regexp.MustCompile(`[^a-z0-9_-]+`)

// resolve (scopeId, profile, etc.)
func SubjToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reBad.ReplaceAllString(s, "_")
	s = strings.Trim(s, "._-")
	if s == "" {
		return "na"
	}
	return s
}
