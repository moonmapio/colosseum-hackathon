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

	return mux
}
