package core

import (
	"fmt"
	"strconv"
	"strings"

	"moonmap.io/go-commons/helpers"
	"moonmap.io/go-commons/persistence"
)

type VariantPlan struct {
	Kind    string
	Width   int
	Height  int
	Quality int
	Ext     string
	Mime    string
	Key     string
}

type VariantsMatrix struct {
	MediaType string
	Key       string
	Urls      map[string]string
	Plan      []VariantPlan
}

func (r *PresignReq) CreateVariantsMatrix(clientBase string) VariantsMatrix {
	matrix := VariantsMatrix{
		Urls: make(map[string]string),
		Plan: make([]VariantPlan, 0, 8),
	}

	mt := helpers.ClassifyMime(r.Mime) // "IMAGE" | "GIF" | "VIDEO" | ...
	ext := helpers.SafeExtFromMime(r.Mime, r.Ext)
	key := r.MakeKeyWithExt(ext)

	matrix.MediaType = mt
	matrix.Key = key
	matrix.Urls["original"] = joinURL(clientBase, key)

	switch mt {
	case "IMAGE":
		r.planImageVariants(&matrix, clientBase)
	case "GIF":
		r.planGifVariants(&matrix, clientBase)
	case "VIDEO":
		r.planVideoVariants(&matrix, clientBase)
	default:
		// deja solo original
	}

	return matrix
}

func (m VariantsMatrix) PlannedForDB() []persistence.PlannedVariantDB {
	out := make([]persistence.PlannedVariantDB, 0, len(m.Plan))
	for _, p := range m.Plan {
		out = append(out, persistence.PlannedVariantDB{
			Kind: p.Kind,
			Key:  p.Key,
			Mime: p.Mime,
			W:    p.Width,
			H:    p.Height,
			Q:    p.Quality,
		})
	}
	return out
}

// ---------- planners por tipo ----------

func (r *PresignReq) planImageVariants(m *VariantsMatrix, clientBase string) {
	sizes := imageSizesForProfile(r.Profile)
	for _, w := range sizes {
		k := VariantKey(m.Key, w, 80, "webp")
		m.Urls[strconv.Itoa(w)] = joinURL(clientBase, k)
		m.Plan = append(m.Plan, VariantPlan{
			Kind:    strconv.Itoa(w),
			Width:   w,
			Height:  0,
			Quality: 80,
			Ext:     "webp",
			Mime:    "image/webp",
			Key:     k,
		})
	}
}

func (r *PresignReq) planGifVariants(m *VariantsMatrix, clientBase string) {
	poster := derivedKey(m.Key, "poster.webp")
	mp4480 := derivedKey(m.Key, "mp4_480.mp4")

	m.Urls["poster"] = joinURL(clientBase, poster)
	m.Urls["mp4_480"] = joinURL(clientBase, mp4480)

	m.Plan = append(m.Plan,
		VariantPlan{Kind: "poster", Width: 512, Height: 0, Quality: 80, Ext: "webp", Mime: "image/webp", Key: poster},
		VariantPlan{Kind: "mp4_480", Width: 0, Height: 0, Quality: 0, Ext: "mp4", Mime: "video/mp4", Key: mp4480},
	)
}

func (r *PresignReq) planVideoVariants(m *VariantsMatrix, clientBase string) {
	poster := derivedKey(m.Key, "poster.webp")
	mp4480 := derivedKey(m.Key, "mp4_480.mp4")
	mp4720 := derivedKey(m.Key, "mp4_720.mp4")

	m.Urls["poster"] = joinURL(clientBase, poster)
	m.Urls["mp4_480"] = joinURL(clientBase, mp4480)
	m.Urls["mp4_720"] = joinURL(clientBase, mp4720)

	m.Plan = append(m.Plan,
		VariantPlan{Kind: "poster", Width: 720, Height: 0, Quality: 80, Ext: "webp", Mime: "image/webp", Key: poster},
		VariantPlan{Kind: "mp4_480", Width: 0, Height: 0, Quality: 0, Ext: "mp4", Mime: "video/mp4", Key: mp4480},
		VariantPlan{Kind: "mp4_720", Width: 0, Height: 0, Quality: 0, Ext: "mp4", Mime: "video/mp4", Key: mp4720},
	)
}

// ---------- helpers de keys/urls ----------

// Convención: <namespace>/<scopeId>/<profile>/<entityId>/v1/original.<ext>
func (r *PresignReq) MakeKey() string {
	return r.MakeKeyWithExt(r.Ext)
}

func (r *PresignReq) MakeKeyWithExt(ext string) string {
	ns := strings.Trim(strings.ToLower(r.Namespace), "/")
	pf := strings.Trim(strings.ToLower(r.Profile), "/")
	id := strings.Trim(r.EntityID, "/")
	sc := strings.Trim(r.ScopeID, "/")
	if ns == "" {
		ns = "media"
	}
	ext = strings.TrimLeft(ext, ".")
	return fmt.Sprintf("%s/%s/%s/%s/v1/original.%s", ns, sc, pf, id, ext)
}

// Para variantes que no dependen de tamaños (poster, mp4_480, etc.)
func derivedKey(originalKey, name string) string {
	i := strings.LastIndex(originalKey, "/")
	if i < 0 {
		return name
	}
	base := originalKey[:i]
	return base + "/" + name
}

// Une base tipo "https://s3.moonmap.io/" con el key
func joinURL(base, key string) string {
	if strings.HasSuffix(base, "/") {
		return base + key
	}
	return base + "/" + key
}

func imageSizesForProfile(profile string) []int {
	switch strings.ToUpper(profile) {
	case "AVATAR":
		return []int{256, 512, 1024}
	case "LOGO":
		return []int{512, 1024}
	case "COVER", "BANNER":
		return []int{1920}
	case "POST_IMAGE":
		return []int{720, 1080}
	default:
		return []int{512, 1024}
	}
}

func VariantKey(key string, w, q int, ext string) string {
	i := strings.LastIndex(key, "/")
	if i < 0 {
		return fmt.Sprintf("%d_q%d.%s", w, q, ext)
	}
	base := key[:i]
	return fmt.Sprintf("%s/%d_q%d.%s", base, w, q, ext)
}
