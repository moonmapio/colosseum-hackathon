package consumer

import (
	"net/http"

	"moonmap.io/go-commons/ownhttp"
)

func (s *Consumer) consumerRoutes() *http.ServeMux {
	mux := ownhttp.Routes()
	return mux
}
