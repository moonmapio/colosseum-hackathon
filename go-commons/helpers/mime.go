package helpers

import "strings"

func SafeExtFromMime(mime, fallback string) string {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	default:
		f := strings.TrimLeft(fallback, ".")
		if f == "" {
			f = "bin"
		}
		return f
	}
}

func ClassifyMime(mime string) string {
	m := strings.ToLower(strings.TrimSpace(mime))
	switch {
	case m == "image/gif":
		return "GIF"
	case strings.HasPrefix(m, "image/"):
		return "IMAGE"
	case strings.HasPrefix(m, "video/"):
		return "VIDEO"
	default:
		return "UNKNOWN"
	}
}
