package ownhttp

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"moonmap.io/go-commons/helpers"
)

func WithCORS(next http.Handler) http.Handler {
	origin := helpers.GetEnv("ALLOW_ORIGIN", "*")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite desde cualquier origen (ojo en prod!)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			// responder rápido a preflight
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// opcional: CORS preflight básico
	mux.HandleFunc("/cors", func(w http.ResponseWriter, r *http.Request) {
		LogRequest(r)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func IsOptionsMethod(r *http.Request, w http.ResponseWriter) bool {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
		return true
	}

	return false
}

func WithLogging(name string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logrus.Infof("%s started", name)

		h(w, r)

		elapsed := time.Since(start)
		logrus.Infof("%s took %s", name, elapsed)
	}
}
