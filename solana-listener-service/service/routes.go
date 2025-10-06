package service

import (
	"net/http"

	"moonmap.io/go-commons/ownhttp"
)

func (s *Service) routes() *http.ServeMux {
	mux := ownhttp.Routes()
	return mux
}
