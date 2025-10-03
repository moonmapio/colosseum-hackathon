package helpers

import (
	"math/big"
	"mime"
	"net/url"
	"path"
	"strings"
)

func ParseUintToBig(s string) *big.Int {
	z := new(big.Int)
	_, _ = z.SetString(s, 10)
	return z
}

func NormCT(ct string) string {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	return ct
}

func ErrStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func DenormalizeChain(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func EnsureNotNilPtr(p **string) {
	if *p == nil {
		empty := ""
		*p = &empty
	}
}

func Strptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func StrFromAny(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func StrFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func EnsureURL(s string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}
	return "https://" + s
}

func MatchAny(lines []string, needles []string) (bool, []string) {
	if len(lines) == 0 || len(needles) == 0 {
		return false, nil
	}
	norm := make([]string, 0, len(needles))
	orig := make([]string, 0, len(needles))
	for _, n := range needles {
		n = strings.TrimSpace(n)
		if n != "" {
			norm = append(norm, strings.ToLower(n))
			orig = append(orig, n)
		}
	}
	seen := make(map[string]struct{}, len(norm))
	var hits []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		ll := strings.ToLower(l)
		for i, n := range norm {
			if strings.Contains(ll, n) {
				if _, ok := seen[orig[i]]; !ok {
					seen[orig[i]] = struct{}{}
					hits = append(hits, orig[i])
				}
			}
		}
	}
	return len(hits) > 0, hits
}

func ContainsCI(haystack, needle string) bool {
	// comparación simple case-insensitive
	h := []rune(haystack)
	n := []rune(needle)
	for i := range h {
		if i+len(n) > len(h) {
			return false
		}
		ok := true
		for j := range n {
			a := h[i+j]
			b := n[j]
			if a >= 'A' && a <= 'Z' {
				a = a - 'A' + 'a'
			}
			if b >= 'A' && b <= 'Z' {
				b = b - 'A' + 'a'
			}
			if a != b {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func FilterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func SliceHasAny(hay []string, needles ...string) bool {
	if len(hay) == 0 {
		return false
	}
	m := make(map[string]struct{}, len(hay))
	for _, h := range hay {
		h = strings.TrimSpace(h)
		if h != "" {
			m[h] = struct{}{}
		}
	}
	for _, n := range needles {
		if _, ok := m[n]; ok {
			return true
		}
	}
	return false
}

func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func SplitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.ToUpper(p))
		}
	}
	return out
}

func UniqueUpper(in []string) []string {
	m := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		su := strings.ToUpper(strings.TrimSpace(s))
		if su == "" {
			continue
		}
		if _, ok := m[su]; ok {
			continue
		}
		m[su] = struct{}{}
		out = append(out, su)
	}
	return out
}

func CoalesceStr(p *string, dflt string) string {
	if p != nil && strings.TrimSpace(*p) != "" {
		return *p
	}
	return dflt
}

func IdToString(v any) string {
	// si es ObjectID
	type oid interface{ Hex() string }
	if x, ok := v.(oid); ok {
		return x.Hex()
	}
	return ""
}

func Matches(pattern, subject string) bool {
	if pattern == subject {
		return true
	}
	// soporte tipo wildcard ">"
	if strings.HasSuffix(pattern, ">") {
		prefix := strings.TrimSuffix(pattern, ">")
		return strings.HasPrefix(subject, prefix)
	}
	return false
}

func IsImageContentType(ct string) bool {
	// maneja "image/png" o "image/png; charset=bla"
	mt, _, _ := mime.ParseMediaType(ct)
	return strings.HasPrefix(mt, "image/")
}

func LooksLikeImageURL(u string) bool {
	pu, err := url.Parse(u)
	if err != nil {
		return false
	}
	ext := strings.ToLower(path.Ext(pu.Path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".svg":
		return true
	default:
		return false
	}
}

func HasText(u string) bool {
	return u != "" && len(u) > 0
}

func ExtractString(m map[string]interface{}, key string) string {
	if m != nil {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}

	return ""
}

func ExtractInt(m map[string]interface{}, key string) (int, bool) {
	if m != nil {
		if v, ok := m[key]; ok {
			switch val := v.(type) {
			case float64: // JSON numbers se decodifican como float64
				return int(val), true
			case int:
				return val, true
			}
		}
	}
	return 0, false
}
