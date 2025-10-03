package helpers

import "strings"

func MatchesAnyEvent(event string, subs []string) bool {
	for _, s := range subs {
		if s == "*" || s == event {
			return true
		}
		// opcional: soportar prefix wildcard tipo "spheres.content.added.*"
		if strings.HasSuffix(s, ".*") {
			prefix := strings.TrimSuffix(s, ".*")
			if strings.HasPrefix(event, prefix) {
				return true
			}
		}
	}
	return false
}
