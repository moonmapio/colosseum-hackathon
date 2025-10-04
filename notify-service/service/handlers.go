package service

import (
	"net/http"

	"moonmap.io/go-commons/ownhttp"
)

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ownhttp.LogRequest(r)
		if ownhttp.IsOptionsMethod(r, w) {
			return
		}
		s.Hub.Add(w, r)
	})

	mux.HandleFunc("/ws/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			ownhttp.WriteJSONError(w, 405, "NOT_ALLOWED", "method")
			return
		}
		stats := s.Hub.Stats()
		ownhttp.WriteJSON(w, 200, map[string]any{
			"stats": stats,
		})
	})

	return mux
}
