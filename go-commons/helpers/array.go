package helpers

import "strings"

func SanitizeStrings(targets []string) []string {
	r := []string{}
	for _, s := range targets {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			continue
		}
		r = append(r, s)
	}

	return r
}
