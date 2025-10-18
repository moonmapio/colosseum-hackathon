package ownhttp

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"moonmap.io/go-commons/helpers"
)

func HeadersFromEnv() http.Header {
	h := http.Header{}

	raw := helpers.GetEnv("HTTP_HEADERS", "")
	if raw != "" {
		var m map[string]string
		if json.Unmarshal([]byte(raw), &m) == nil {
			for k, v := range m {
				if k != "" && v != "" {
					h.Set(k, v)
				}
			}
		} else {
			parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ';' || r == ',' || r == '|' })
			for _, p := range parts {
				kv := strings.SplitN(p, "=", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(kv[0])
					v := strings.TrimSpace(kv[1])
					if k != "" && v != "" {
						h.Set(k, v)
					}
				}
			}
		}
	}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "HEADER_") {
			continue
		}
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimPrefix(kv[0], "HEADER_")
		key = strings.ReplaceAll(key, "_", "-")
		val := kv[1]
		if key != "" && val != "" {
			h.Set(key, val)
		}
	}

	return h
}
